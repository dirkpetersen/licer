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
	"os"
	"strings"
)

func CanRemoveHeader(filename string, config *Config) (bool, error) {
	// First, check if there's a header with SPDX identifier
	headerInfo, err := DetectExistingHeader(filename)
	if err != nil {
		return false, err
	}
	
	if !headerInfo.HasHeader {
		return false, nil // No header to remove
	}
	
	// Read the header content to check ownership
	content, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}
	
	lines := strings.Split(string(content), "\n")
	
	// Extract header lines
	var headerLines []string
	start := headerInfo.StartLine
	end := headerInfo.EndLine
	
	if start < len(lines) && end < len(lines) {
		headerLines = lines[start:end+1]
	}
	
	headerText := strings.Join(headerLines, "\n")
	headerLower := strings.ToLower(headerText)
	
	// Check for SPDX identifier (case-insensitive)
	hasSPDX := strings.Contains(headerLower, "spdx-license-identifier")
	if !hasSPDX {
		return false, nil // No SPDX identifier, not safe to remove
	}
	
	// Check ownership - must contain user's name OR organization name
	hasUserName := strings.Contains(headerText, config.FullName)
	hasOrgName := strings.Contains(headerText, config.Organization)
	
	return hasUserName || hasOrgName, nil
}

func RemoveHeader(filename string) error {
	// Detect the header
	headerInfo, err := DetectExistingHeader(filename)
	if err != nil {
		return err
	}
	
	if !headerInfo.HasHeader {
		return nil // Nothing to remove
	}
	
	// Read the entire file
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	
	lines := strings.Split(string(content), "\n")
	var newContent []string
	
	if headerInfo.HasShebang {
		// Keep shebang, remove header after it
		newContent = append(newContent, lines[0]) // Keep shebang
		
		// Skip header lines and any blank lines immediately following
		skipIndex := headerInfo.EndLine + 1
		for skipIndex < len(lines) && strings.TrimSpace(lines[skipIndex]) == "" {
			skipIndex++
		}
		
		// Add remaining content
		if skipIndex < len(lines) {
			newContent = append(newContent, lines[skipIndex:]...)
		}
	} else {
		// Remove header from beginning
		skipIndex := headerInfo.EndLine + 1
		
		// Skip any blank lines immediately following the header
		for skipIndex < len(lines) && strings.TrimSpace(lines[skipIndex]) == "" {
			skipIndex++
		}
		
		// Add remaining content
		if skipIndex < len(lines) {
			newContent = append(newContent, lines[skipIndex:]...)
		}
	}
	
	// Write the modified content back
	newContentStr := strings.Join(newContent, "\n")
	return os.WriteFile(filename, []byte(newContentStr), 0644)
}