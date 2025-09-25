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
	"path/filepath"
	"sync"
	"sync/atomic"
)

type Crawler struct {
	config      *Config
	forceReplace bool
	removeMode  bool
	verbose     bool
	stats       *ProcessingStats
}

type ProcessingStats struct {
	FilesProcessed int64
	FilesModified  int64
	FilesSkipped   int64
	FilesErrored   int64
}

func NewCrawler(config *Config, forceReplace, removeMode, verbose bool) *Crawler {
	return &Crawler{
		config:      config,
		forceReplace: forceReplace,
		removeMode:  removeMode,
		verbose:     verbose,
		stats:       &ProcessingStats{},
	}
}

func (c *Crawler) ProcessRepository(repoRoot string) error {
	if c.verbose {
		fmt.Printf("Starting parallel processing of repository: %s\n", repoRoot)
	}
	
	// Manage LICENSE file first (only if not in remove mode)
	if !c.removeMode {
		err := ManageLicenseFile(repoRoot, c.config, c.verbose)
		if err != nil {
			if c.verbose {
				fmt.Printf("[LICENSE] Error managing LICENSE file: %v\n", err)
			}
		}
	}
	
	err := c.processDirectoryRecursive(repoRoot)
	if err != nil {
		return err
	}
	
	if c.verbose {
		c.printStats()
	}
	
	return nil
}

func (c *Crawler) processDirectoryRecursive(dir string) error {
	// Check if this is the .git directory (skip it)
	if filepath.Base(dir) == ".git" {
		return nil
	}
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		if c.verbose {
			fmt.Printf("[ERROR] Failed to read directory %s: %v\n", dir, err)
		}
		return nil // Don't fail completely, just skip this directory
	}
	
	var wg sync.WaitGroup
	
	// Process files in current directory sequentially
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		filename := filepath.Join(dir, entry.Name())
		result := ProcessFile(filename, c.config, c.forceReplace, c.removeMode, false) // Don't log here to avoid race conditions
		
		// Update statistics
		atomic.AddInt64(&c.stats.FilesProcessed, 1)
		if result.Modified {
			atomic.AddInt64(&c.stats.FilesModified, 1)
		} else if result.Action == "SKIP" {
			atomic.AddInt64(&c.stats.FilesSkipped, 1)
		}
		
		// Log result in thread-safe way
		if c.verbose {
			c.logResultSafe(filename, result)
		}
	}
	
	// Launch workers for subdirectories with per-directory concurrency limit
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent subdirs per directory level
	
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".git" {
			continue
		}
		
		wg.Add(1)
		go func(subdirName string) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			subdirPath := filepath.Join(dir, subdirName)
			if err := c.processDirectoryRecursive(subdirPath); err != nil {
				if c.verbose {
					fmt.Printf("[ERROR] Failed processing directory %s: %v\n", subdirPath, err)
				}
			}
		}(entry.Name())
	}
	
	// Wait for all subdirectory workers to complete
	wg.Wait()
	return nil
}

var logMutex sync.Mutex

func (c *Crawler) logResultSafe(filename string, result ProcessResult) {
	logMutex.Lock()
	defer logMutex.Unlock()
	LogResult(filename, result, true)
}

func (c *Crawler) printStats() {
	fmt.Printf("\n=== Processing Summary ===\n")
	fmt.Printf("Files processed: %d\n", c.stats.FilesProcessed)
	fmt.Printf("Files modified:  %d\n", c.stats.FilesModified)
	fmt.Printf("Files skipped:   %d\n", c.stats.FilesSkipped)
	fmt.Printf("=========================\n")
}