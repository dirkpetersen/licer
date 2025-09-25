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
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	gitFolder string
	force     bool
	remove    bool
	hook      bool
	preCommit bool
	verbose   bool
	help      bool
)

func init() {
	flag.StringVar(&gitFolder, "git-folder", "", "Path to git repository (default: current directory)")
	flag.BoolVar(&force, "force", false, "Force replacement of existing headers")
	flag.BoolVar(&remove, "remove", false, "Remove existing headers (requires SPDX-License-Identifier and ownership match)")
	flag.BoolVar(&hook, "hook", false, "Install/uninstall Git pre-commit hook")
	flag.BoolVar(&preCommit, "pre-commit", false, "Pre-commit mode: process only newly staged files")
	flag.BoolVar(&verbose, "verbose", true, "Verbose output")
	flag.BoolVar(&help, "help", false, "Show help message")
}

func main() {
	flag.Parse()
	
	if help {
		printUsage()
		return
	}

	// Validate mutually exclusive flags
	if force && remove {
		log.Fatalf("--force and --remove cannot be used together")
	}
	
	// Handle hook management mode
	if hook {
		handleHookManagement(remove, verbose)
		return
	}
	
	// Handle pre-commit mode
	if preCommit {
		handlePreCommitMode()
		return
	}

	// Determine the git repository root
	repoRoot := gitFolder
	if repoRoot == "" {
		var err error
		repoRoot, err = os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current directory: %v", err)
		}
	}

	// Convert to absolute path
	absRepoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Verify it's a git repository
	gitDir := filepath.Join(absRepoRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		log.Fatalf("Not a git repository: %s", absRepoRoot)
	}

	if verbose {
		fmt.Printf("Licer - License Header Management Tool\n")
		fmt.Printf("Working in git repository: %s\n", absRepoRoot)
		fmt.Printf("Force mode: %v\n", force)
		fmt.Printf("Remove mode: %v\n", remove)
		fmt.Printf("Verbose mode: %v\n", verbose)
		fmt.Println()
	}

	// Load or create configuration
	config, err := LoadOrCreateConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if verbose {
		fmt.Printf("Configuration:\n")
		fmt.Printf("  Name: %s\n", config.FullName)
		fmt.Printf("  Role: %s\n", config.DefaultRole)
		fmt.Printf("  Department/Lab: %s\n", config.DeptOrLab)
		fmt.Printf("  Organization: %s\n", config.Organization)
		
		template := GetHeaderTemplate(config)
		fmt.Printf("  License: %s\n", template.LicenseType)
		fmt.Printf("  Copyright Owner: %s\n", template.CopyrightOwner)
		fmt.Println()
	}

	// Check for hook installation prompt (only if no git-folder specified)
	if gitFolder == "" && !isHookInstalled(absRepoRoot) {
		if promptForHookInstallation() {
			if err := installPreCommitHook(absRepoRoot, verbose); err != nil {
				fmt.Printf("Warning: Failed to install hook: %v\n", err)
			}
		}
	}

	// Start crawling and processing
	crawler := NewCrawler(config, force, remove, verbose)
	if err := crawler.ProcessRepository(absRepoRoot); err != nil {
		log.Fatalf("Failed to process repository: %v", err)
	}

	if verbose {
		fmt.Println("Processing completed successfully!")
	}
}

func printUsage() {
	fmt.Println("Licer - License Header Management Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  licer [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Description:")
	fmt.Println("  Licer recursively crawls a git repository and adds copyright headers")
	fmt.Println("  to source files based on your role configuration.")
	fmt.Println()
	fmt.Println("  On first run, you'll be prompted to create a configuration file at")
	fmt.Println("  ~/.config/licer.yml with your name, role, department, and organization.")
	fmt.Println()
	fmt.Println("  Students get MIT license headers, Faculty/Staff get Apache 2.0 headers.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  licer                                # Process current git repository")
	fmt.Println("  licer --git-folder /path/to/repo     # Process specific repository")
	fmt.Println("  licer --force                        # Replace existing headers")
	fmt.Println("  licer --remove                       # Remove existing headers (safe mode)")
	fmt.Println("  licer --hook                         # Install Git pre-commit hook")
	fmt.Println("  licer --hook --remove                # Uninstall pre-commit hook")
	fmt.Println("  licer --verbose=false                # Quiet mode")
}