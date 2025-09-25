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
	"os"
	"strings"
)

type HeaderInfo struct {
	HasHeader         bool
	HasThirdPartyCopyright bool
	StartLine         int
	EndLine           int
	HasShebang        bool
}

func DetectExistingHeader(filename string) (HeaderInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return HeaderInfo{}, err
	}
	defer file.Close()
	
	info := HeaderInfo{
		HasHeader:              false,
		HasThirdPartyCopyright: false,
		StartLine:              -1,
		EndLine:                -1,
		HasShebang:             false,
	}
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	maxLinesToCheck := 20
	
	// Read first few lines to check for shebang and third-party copyright
	var firstThreeLines []string
	
	// Check first line for shebang
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		firstThreeLines = append(firstThreeLines, line)
		lineNum++
		
		if strings.HasPrefix(line, "#!") {
			info.HasShebang = true
		}
		
		// Check for SPDX identifier in first line (rare but possible)
		if containsSPDXIdentifier(line) {
			info.HasHeader = true
			info.StartLine = lineNum - 1 // 0-based
		}
	}
	
	// Read next two lines for third-party copyright detection
	for i := 0; i < 2 && scanner.Scan(); i++ {
		line := strings.TrimSpace(scanner.Text())
		firstThreeLines = append(firstThreeLines, line)
		lineNum++
		
		if containsSPDXIdentifier(line) {
			info.HasHeader = true
			if info.StartLine == -1 {
				info.StartLine = findHeaderStart(filename, lineNum)
			}
			info.EndLine = lineNum - 1 // 0-based, this line contains SPDX
		}
	}
	
	// Check for third-party copyright in first 3 lines (excluding SPDX headers)
	if !info.HasHeader {
		for _, line := range firstThreeLines {
			if strings.Contains(strings.ToLower(line), "copyright") {
				info.HasThirdPartyCopyright = true
				break
			}
		}
	}
	
	// Continue scanning for SPDX identifier in remaining lines
	for scanner.Scan() && lineNum < maxLinesToCheck {
		line := strings.TrimSpace(scanner.Text())
		lineNum++
		
		if containsSPDXIdentifier(line) {
			info.HasHeader = true
			if info.StartLine == -1 {
				// Find the start of the header block
				info.StartLine = findHeaderStart(filename, lineNum)
			}
			info.EndLine = lineNum - 1 // 0-based, this line contains SPDX
			break
		}
	}
	
	// If we found a header, extend the end to include any following copyright/license lines
	if info.HasHeader {
		info.EndLine = findHeaderEnd(filename, info.EndLine)
	} else if info.HasThirdPartyCopyright {
		// For third-party copyright, find the end of the license block
		info.StartLine, info.EndLine = findThirdPartyCopyrightBlock(filename)
	}
	
	return info, scanner.Err()
}

func containsSPDXIdentifier(line string) bool {
	return strings.Contains(strings.ToLower(line), "spdx-license-identifier")
}

func findHeaderStart(filename string, spdxLine int) int {
	file, err := os.Open(filename)
	if err != nil {
		return 0
	}
	defer file.Close()
	
	lines := make([]string, 0, spdxLine)
	scanner := bufio.NewScanner(file)
	
	// Read lines up to SPDX line
	for i := 0; i < spdxLine && scanner.Scan(); i++ {
		lines = append(lines, scanner.Text())
	}
	
	// Work backwards from SPDX line to find start of header
	startLine := 0
	
	// Skip shebang if present
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "#!") {
		startLine = 1
	}
	
	// Look for copyright notice or other header indicators
	for i := spdxLine - 2; i >= startLine; i-- { // spdxLine is 1-based, array is 0-based
		if i < 0 || i >= len(lines) {
			continue
		}
		
		line := strings.ToLower(strings.TrimSpace(lines[i]))
		
		if strings.Contains(line, "copyright") ||
		   strings.Contains(line, "licensed under") ||
		   strings.Contains(line, "developed by") ||
		   strings.Contains(line, "author") ||
		   isCommentLine(lines[i]) {
			continue
		} else {
			// Found non-header line, start is after this
			return i + 1
		}
	}
	
	return startLine
}

func findHeaderEnd(filename string, spdxLine int) int {
	file, err := os.Open(filename)
	if err != nil {
		return spdxLine
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	// Skip to SPDX line
	for lineNum <= spdxLine && scanner.Scan() {
		lineNum++
	}
	
	endLine := spdxLine
	
	// Continue scanning for related header content
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" && isCommentLine(scanner.Text()) {
			// Empty comment line, might be part of header
			endLine = lineNum - 1
			continue
		}
		
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "see license") ||
		   strings.Contains(lowerLine, "developed by") ||
		   strings.Contains(lowerLine, "oregon state university") ||
		   isCommentLine(scanner.Text()) {
			endLine = lineNum - 1
		} else {
			// Found non-header content
			break
		}
	}
	
	return endLine
}

func isCommentLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	
	// Check common comment prefixes
	commentPrefixes := []string{"//", "#", "/*", "*", ";;", "--", "\"", "REM", "C", "!", "%"}
	
	for _, prefix := range commentPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	
	return false
}

func findThirdPartyCopyrightBlock(filename string) (int, int) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, 0
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	startLine := -1
	endLine := -1
	
	// Skip shebang if present
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++
		
		if strings.HasPrefix(line, "#!") {
			// Shebang found, start looking from next line
		} else if strings.Contains(strings.ToLower(line), "copyright") {
			startLine = lineNum - 1 // 0-based
		}
	}
	
	// Continue scanning for copyright or license-related content
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++
		
		lineLower := strings.ToLower(line)
		
		// If we haven't found the start yet, look for copyright
		if startLine == -1 && strings.Contains(lineLower, "copyright") {
			startLine = lineNum - 1 // 0-based
		}
		
		// If we have a start, look for the end of license text
		if startLine != -1 {
			// Consider it part of the license block if it contains license-related keywords
			if strings.Contains(lineLower, "permission") ||
			   strings.Contains(lineLower, "license") ||
			   strings.Contains(lineLower, "software") ||
			   strings.Contains(lineLower, "rights") ||
			   strings.Contains(lineLower, "distribute") ||
			   strings.Contains(lineLower, "modify") ||
			   strings.Contains(lineLower, "use") ||
			   strings.Contains(lineLower, "without warranty") ||
			   strings.Contains(lineLower, "liability") ||
			   strings.Contains(lineLower, "damages") ||
			   isCommentLine(scanner.Text()) ||
			   line == "" {
				endLine = lineNum - 1 // 0-based, continue expanding
			} else {
				// Found non-license content, end the block
				break
			}
		}
	}
	
	// If we found a start but no clear end, assume it goes to the end of license text we saw
	if startLine != -1 && endLine == -1 {
		endLine = startLine // Minimal block
	}
	
	return startLine, endLine
}

func HasShebang(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		firstLine := strings.TrimSpace(scanner.Text())
		return strings.HasPrefix(firstLine, "#!"), nil
	}
	
	return false, scanner.Err()
}