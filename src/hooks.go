// Copyright 2025 Oregon State University
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file for details.
// SPDX-License-Identifier: Apache-2.0
//
// Developed by: Dirk Petersen
//               UIT/ARCS

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const preCommitHookScript = `#!/bin/bash

# Licer pre-commit hook - Automatically add license headers to new files

# Get the directory where licer binary is located
LICER_PATH="$(which licer)"
if [ -z "$LICER_PATH" ]; then
    # Try to find licer in common locations
    REPO_ROOT="$(git rev-parse --show-toplevel)"
    for path in "./licer" "../licer" "$REPO_ROOT/licer"; do
        if [ -x "$path" ]; then
            LICER_PATH="$path"
            break
        fi
    done
fi

if [ -z "$LICER_PATH" ]; then
    echo "Warning: licer not found, skipping header check" >&2
    exit 0
fi

# Run licer in pre-commit mode
"$LICER_PATH" --pre-commit --verbose=false

exit 0
`

func handleHookManagement(removeMode bool, verbose bool) {
	repoRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	
	// Verify it's a git repository
	gitDir := filepath.Join(repoRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		log.Fatalf("Not a git repository: %s", repoRoot)
	}
	
	if removeMode {
		err := uninstallPreCommitHook(repoRoot, verbose)
		if err != nil {
			log.Fatalf("Failed to uninstall hook: %v", err)
		}
		if verbose {
			fmt.Println("Pre-commit hook uninstalled successfully")
		}
	} else {
		err := installPreCommitHook(repoRoot, verbose)
		if err != nil {
			log.Fatalf("Failed to install hook: %v", err)
		}
		if verbose {
			fmt.Println("Pre-commit hook installed successfully")
		}
	}
}

func handlePreCommitMode() {
	// Get current working directory (should be repo root when called by git)
	repoRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	
	// Load configuration
	config, err := LoadOrCreateConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	
	// Get newly staged files
	newFiles, err := getStagedNewFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting staged files: %v\n", err)
		os.Exit(1)
	}
	
	if len(newFiles) == 0 {
		// No new files to process
		os.Exit(0)
	}
	
	// Process each new file
	hasErrors := false
	for _, filename := range newFiles {
		fullPath := filepath.Join(repoRoot, filename)
		
		// Check if file exists (might have been deleted after staging)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}
		
		result := ProcessFile(fullPath, config, false, false, false) // Never force in pre-commit mode
		if result.Modified {
			// Re-stage the modified file
			cmd := exec.Command("git", "add", filename)
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error re-staging %s: %v\n", filename, err)
				hasErrors = true
			}
		}
	}
	
	if hasErrors {
		os.Exit(1)
	}
	os.Exit(0)
}

func getStagedNewFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}
	
	var newFiles []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		// Format: "A\tfilename" for added files
		if strings.HasPrefix(line, "A\t") {
			filename := strings.TrimPrefix(line, "A\t")
			newFiles = append(newFiles, filename)
		}
	}
	
	return newFiles, nil
}

func isHookInstalled(repoRoot string) bool {
	hookPath := filepath.Join(repoRoot, ".git", "hooks", "pre-commit")
	
	// Check if hook file exists and is executable
	info, err := os.Stat(hookPath)
	if os.IsNotExist(err) {
		return false
	}
	
	if info.Mode()&0111 == 0 {
		return false // Not executable
	}
	
	// Check if it contains licer integration
	content, err := os.ReadFile(hookPath)
	if err != nil {
		return false
	}
	
	return strings.Contains(string(content), "licer --pre-commit")
}

func installPreCommitHook(repoRoot string, verbose bool) error {
	hooksDir := filepath.Join(repoRoot, ".git", "hooks")
	hookPath := filepath.Join(hooksDir, "pre-commit")
	backupPath := filepath.Join(hooksDir, "pre-commit.backup")
	
	// Create hooks directory if it doesn't exist
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}
	
	// Backup existing hook if it exists
	if _, err := os.Stat(hookPath); err == nil {
		if verbose {
			fmt.Printf("Backing up existing pre-commit hook to pre-commit.backup\n")
		}
		if err := os.Rename(hookPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup existing hook: %w", err)
		}
	}
	
	// Write new hook
	if err := os.WriteFile(hookPath, []byte(preCommitHookScript), 0755); err != nil {
		return fmt.Errorf("failed to write hook script: %w", err)
	}
	
	if verbose {
		fmt.Printf("Pre-commit hook installed at %s\n", hookPath)
	}
	
	return nil
}

func uninstallPreCommitHook(repoRoot string, verbose bool) error {
	hookPath := filepath.Join(repoRoot, ".git", "hooks", "pre-commit")
	backupPath := filepath.Join(repoRoot, ".git", "hooks", "pre-commit.backup")
	
	// Check if our hook is installed
	if !isHookInstalled(repoRoot) {
		if verbose {
			fmt.Println("No licer pre-commit hook found to uninstall")
		}
		return nil
	}
	
	// Remove the hook
	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("failed to remove hook: %w", err)
	}
	
	// Restore backup if it exists
	if _, err := os.Stat(backupPath); err == nil {
		if verbose {
			fmt.Printf("Restoring backed up pre-commit hook\n")
		}
		if err := os.Rename(backupPath, hookPath); err != nil {
			return fmt.Errorf("failed to restore backup hook: %w", err)
		}
	}
	
	if verbose {
		fmt.Printf("Pre-commit hook uninstalled\n")
	}
	
	return nil
}

func promptForHookInstallation() bool {
	fmt.Print("Install pre-commit hook to automatically license new files? (y/N): ")
	
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}