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
	"path/filepath"
	"strings"
	"unicode"
)

type CommentStyle struct {
	Line       string
	BlockStart string
	BlockEnd   string
}

var commentStyles = map[string]CommentStyle{
	".go":    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".py":    {Line: "#"},
	".sh":    {Line: "#"},
	".rb":    {Line: "#"},
	".js":    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".mjs":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".cjs":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".ts":    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".tsx":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".jsx":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".html":  {Line: "<!--", BlockStart: "<!--", BlockEnd: "-->"},
	".htm":   {Line: "<!--", BlockStart: "<!--", BlockEnd: "-->"},
	".css":   {Line: "/*", BlockStart: "/*", BlockEnd: "*/"},
	".scss":  {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".sass":  {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".less":  {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".java":  {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".c":     {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".cpp":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".cc":    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".cxx":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".h":     {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".hpp":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".rs":    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".swift": {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".kt":    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".scala": {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".cs":    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".yaml":  {Line: "#"},
	".yml":   {Line: "#"},
	".toml":  {Line: "#"},
	".ini":   {Line: "#"},
	".cfg":   {Line: "#"},
	".conf":  {Line: "#"},
	".sql":   {Line: "--", BlockStart: "/*", BlockEnd: "*/"},
	".lua":   {Line: "--", BlockStart: "--[[", BlockEnd: "--]]"},
	".r":     {Line: "#"},
	".R":     {Line: "#"},
	".rmd":   {Line: "<!--", BlockStart: "<!--", BlockEnd: "-->"},
	".Rmd":   {Line: "<!--", BlockStart: "<!--", BlockEnd: "-->"},
	".m":     {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".mm":    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".vim":   {Line: "\""},
	".vimrc": {Line: "\""},
	".el":    {Line: ";;"},
	".lisp":  {Line: ";;"},
	".lsp":   {Line: ";;"},
	".clj":   {Line: ";;"},
	".cljs":  {Line: ";;"},
	".hs":    {Line: "--", BlockStart: "{-", BlockEnd: "-}"},
	".lhs":   {Line: "--", BlockStart: "{-", BlockEnd: "-}"},
	".ml":    {Line: "(*", BlockEnd: "*)"},
	".mli":   {Line: "(*", BlockEnd: "*)"},
	".pas":   {Line: "//", BlockStart: "{", BlockEnd: "}"},
	".pl":    {Line: "#"},
	".pm":    {Line: "#"},
	".php":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".dart":  {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".f":     {Line: "C", BlockStart: "C", BlockEnd: "C"},
	".f90":   {Line: "!", BlockStart: "!", BlockEnd: "!"},
	".f95":   {Line: "!", BlockStart: "!", BlockEnd: "!"},
	".jl":    {Line: "#", BlockStart: "#=", BlockEnd: "=#"},
	".zig":   {Line: "//"},
	".nim":   {Line: "#", BlockStart: "#[", BlockEnd: "]#"},
	".cr":    {Line: "#"},
	".d":     {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".ex":    {Line: "#"},
	".exs":   {Line: "#"},
	".erl":   {Line: "%"},
	".hrl":   {Line: "%"},
	".fs":    {Line: "//", BlockStart: "(*", BlockEnd: "*)"},
	".fsx":   {Line: "//", BlockStart: "(*", BlockEnd: "*)"},
	".fsi":   {Line: "//", BlockStart: "(*", BlockEnd: "*)"},
	".v":     {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".vv":    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	".bat":   {Line: "REM"},
	".cmd":   {Line: "REM"},
	".ps1":   {Line: "#", BlockStart: "<#", BlockEnd: "#>"},
	".psm1":  {Line: "#", BlockStart: "<#", BlockEnd: "#>"},
	"":       {Line: "#"}, // No extension = shell script
}

var excludedExtensions = map[string]bool{
	".md":     true,
	".txt":    true,
	".json":   true,
	".xml":    true,
	".csv":    true,
	".tsv":    true,
	".log":    true,
	".out":    true,
	".pdf":    true,
	".doc":    true,
	".docx":   true,
	".xls":    true,
	".xlsx":   true,
	".ppt":    true,
	".pptx":   true,
	".zip":    true,
	".tar":    true,
	".gz":     true,
	".bz2":    true,
	".xz":     true,
	".7z":     true,
	".rar":    true,
	".png":    true,
	".jpg":    true,
	".jpeg":   true,
	".gif":    true,
	".bmp":    true,
	".tiff":   true,
	".svg":    true,
	".ico":    true,
	".mp3":    true,
	".mp4":    true,
	".avi":    true,
	".mov":    true,
	".mkv":    true,
	".wav":    true,
	".flac":   true,
	".exe":    true,
	".dll":    true,
	".so":     true,
	".dylib":  true,
	".a":      true,
	".lib":    true,
	".obj":    true,
	".o":      true,
	".class":  true,
	".jar":    true,
	".war":    true,
	".ear":    true,
	".pyc":    true,
	".pyo":    true,
	".pyd":    true,
	".whl":    true,
	".egg":    true,
	".deb":    true,
	".rpm":    true,
	".msi":    true,
	".dmg":    true,
	".iso":    true,
	".img":    true,
}

func GetCommentStyle(filename string) (CommentStyle, bool) {
	ext := strings.ToLower(filepath.Ext(filename))
	
	// Check if file should be excluded
	if excludedExtensions[ext] {
		return CommentStyle{}, false
	}
	
	// Get comment style
	style, exists := commentStyles[ext]
	if !exists {
		// Check if it might be a text file (no extension)
		if ext == "" {
			if isTextFile(filename) {
				return commentStyles[""], true
			}
			return CommentStyle{}, false
		}
		return CommentStyle{}, false
	}
	
	return style, true
}

func ShouldProcessFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	
	// Skip excluded extensions
	if excludedExtensions[ext] {
		return false
	}
	
	// Skip if no comment style available
	_, exists := commentStyles[ext]
	if !exists && ext != "" {
		return false
	}
	
	// For files with no extension, check if they're text files
	if ext == "" {
		return isTextFile(filename)
	}
	
	return true
}

func isTextFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()
	
	// Read first 512 bytes to check for binary content
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return false
	}
	
	// Check for null bytes or too many non-printable characters
	nullBytes := 0
	nonPrintable := 0
	
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			nullBytes++
		} else if !unicode.IsPrint(rune(buffer[i])) && !unicode.IsSpace(rune(buffer[i])) {
			nonPrintable++
		}
	}
	
	// If more than 30% non-printable or any null bytes, likely binary
	if nullBytes > 0 || float64(nonPrintable)/float64(n) > 0.30 {
		return false
	}
	
	return true
}

func FormatHeader(header string, style CommentStyle) string {
	lines := strings.Split(header, "\n")
	var result []string
	
	// For CSS files, use block comments for better formatting
	if style.Line == "/*" && style.BlockStart == "/*" && style.BlockEnd == "*/" {
		result = append(result, "/*")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				result = append(result, " *")
			} else {
				result = append(result, " * "+line)
			}
		}
		result = append(result, " */")
		return strings.Join(result, "\n")
	}
	
	// Use line comments for headers (more consistent)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, style.Line)
		} else {
			result = append(result, style.Line+" "+line)
		}
	}
	
	return strings.Join(result, "\n")
}