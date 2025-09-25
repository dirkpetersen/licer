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
	"fmt"
	"os"
	"strings"
)

type ProcessResult struct {
	Action   string // "ADD", "REPLACE", "SKIP"
	Reason   string
	Modified bool
}

func ProcessFile(filename string, config *Config, forceReplace bool, removeMode bool, verbose bool) ProcessResult {
	// Handle remove mode
	if removeMode {
		return processRemoveMode(filename, config)
	}
	
	// Check if we should process this file type
	if !ShouldProcessFile(filename) {
		return ProcessResult{
			Action: "SKIP",
			Reason: "Excluded file type",
		}
	}
	
	// Get comment style for this file
	commentStyle, ok := GetCommentStyle(filename)
	if !ok {
		return ProcessResult{
			Action: "SKIP", 
			Reason: "No comment style available",
		}
	}
	
	// Detect existing header
	headerInfo, err := DetectExistingHeader(filename)
	if err != nil {
		return ProcessResult{
			Action: "SKIP",
			Reason: fmt.Sprintf("Error reading file: %v", err),
		}
	}
	
	// Check if file already has header and we're not forcing
	if headerInfo.HasHeader && !forceReplace {
		return ProcessResult{
			Action: "SKIP",
			Reason: "Header already exists",
		}
	}
	
	// Check for third-party copyright - only overwrite with --force
	if headerInfo.HasThirdPartyCopyright && !forceReplace {
		return ProcessResult{
			Action: "SKIP",
			Reason: "Third-party copyright found (use --force to overwrite)",
		}
	}
	
	// Generate new header
	headerText := GenerateHeader(config)
	formattedHeader := FormatHeader(headerText, commentStyle)
	
	// Process the file
	action := "ADD"
	if headerInfo.HasHeader {
		action = "REPLACE"
	} else if headerInfo.HasThirdPartyCopyright {
		action = "REPLACE"
	}
	
	err = modifyFile(filename, formattedHeader, headerInfo)
	if err != nil {
		return ProcessResult{
			Action: "SKIP",
			Reason: fmt.Sprintf("Error modifying file: %v", err),
		}
	}
	
	reason := fmt.Sprintf("Added %s header", GetLicenseType(config))
	if headerInfo.HasThirdPartyCopyright {
		reason = fmt.Sprintf("Replaced third-party copyright with %s header", GetLicenseType(config))
	}
	
	return ProcessResult{
		Action:   action,
		Reason:   reason,
		Modified: true,
	}
}

func modifyFile(filename, newHeader string, headerInfo HeaderInfo) error {
	// Read the entire file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	lines := strings.Split(string(content), "\n")
	
	var newContent []string
	
	if headerInfo.HasHeader || headerInfo.HasThirdPartyCopyright {
		// Replace existing header or third-party copyright
		if headerInfo.HasShebang {
			// Keep shebang
			newContent = append(newContent, lines[0])
			newContent = append(newContent, "")
			newContent = append(newContent, strings.Split(newHeader, "\n")...)
			newContent = append(newContent, "")
			
			// Add remaining content after old header
			if headerInfo.EndLine+1 < len(lines) {
				newContent = append(newContent, lines[headerInfo.EndLine+1:]...)
			}
		} else {
			// Replace from start
			newContent = append(newContent, strings.Split(newHeader, "\n")...)
			newContent = append(newContent, "")
			
			// Add remaining content after old header
			if headerInfo.EndLine+1 < len(lines) {
				newContent = append(newContent, lines[headerInfo.EndLine+1:]...)
			}
		}
	} else {
		// Add new header
		if headerInfo.HasShebang {
			// Keep shebang, add header after
			newContent = append(newContent, lines[0])
			newContent = append(newContent, "")
			newContent = append(newContent, strings.Split(newHeader, "\n")...)
			newContent = append(newContent, "")
			
			// Add rest of original content
			if len(lines) > 1 {
				newContent = append(newContent, lines[1:]...)
			}
		} else {
			// Add header at beginning
			newContent = append(newContent, strings.Split(newHeader, "\n")...)
			newContent = append(newContent, "")
			
			// Add original content
			newContent = append(newContent, lines...)
		}
	}
	
	// Write the modified content back
	newContentStr := strings.Join(newContent, "\n")
	err = os.WriteFile(filename, []byte(newContentStr), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

func GetLicenseType(config *Config) string {
	template := GetHeaderTemplate(config)
	return template.LicenseType
}

func processRemoveMode(filename string, config *Config) ProcessResult {
	// Check if we should process this file type
	if !ShouldProcessFile(filename) {
		return ProcessResult{
			Action: "SKIP",
			Reason: "Excluded file type",
		}
	}
	
	// Check if we can safely remove the header
	canRemove, err := CanRemoveHeader(filename, config)
	if err != nil {
		return ProcessResult{
			Action: "SKIP",
			Reason: fmt.Sprintf("Error checking header: %v", err),
		}
	}
	
	if !canRemove {
		// Check if there's a header at all
		headerInfo, err := DetectExistingHeader(filename)
		if err != nil {
			return ProcessResult{
				Action: "SKIP",
				Reason: fmt.Sprintf("Error reading file: %v", err),
			}
		}
		
		if !headerInfo.HasHeader {
			return ProcessResult{
				Action: "SKIP",
				Reason: "No header found",
			}
		}
		
		return ProcessResult{
			Action: "SKIP",
			Reason: "Header ownership mismatch (safety check)",
		}
	}
	
	// Remove the header
	err = RemoveHeader(filename)
	if err != nil {
		return ProcessResult{
			Action: "SKIP",
			Reason: fmt.Sprintf("Error removing header: %v", err),
		}
	}
	
	return ProcessResult{
		Action:   "REMOVE",
		Reason:   "Removed header (ownership match)",
		Modified: true,
	}
}

func LogResult(filename string, result ProcessResult, verbose bool) {
	if !verbose {
		return
	}
	
	switch result.Action {
	case "ADD":
		fmt.Printf("[ADD] %s - %s\n", filename, result.Reason)
	case "REPLACE":
		fmt.Printf("[REPLACE] %s - %s\n", filename, result.Reason)  
	case "REMOVE":
		fmt.Printf("[REMOVE] %s - %s\n", filename, result.Reason)
	case "SKIP":
		fmt.Printf("[SKIP] %s - %s\n", filename, result.Reason)
	}
}