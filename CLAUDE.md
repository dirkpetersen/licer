# CLAUDE.md requirements for Licer 

Licer is a tool that crawls github repositories recursively and adds copyright headers
to each text file, including source 

## Scope 

- Implement all rules from file POLICY.md 

## exemptions / cautions 

 - do not make changes to these file types: md, json, csv and other file types that 
   would be destroyed by injecting headers 
 - make changes to text files only 
 - if a file has no extension and is a text file, assume that the # is used for comment 
 - only make changes to files that have no headers (idempotency) except when --force is 
   used. In that case replace the entire header  
 - for scripts make sure that you enter the header below the shebang 
 - do not make any changes in the .git subfolder 

## execution

 - will only work if the current directory is the root of a git repository, 
   except when the `--git-folder <git-folder>` argument is used 
 - be verbose about the changes made 

## configuration 

 - will use config file ~/.config/licer.yml and creates it if it does not exist
 - it will contain the variables FULL_NAME, DEFAULT_ROLE, DEPT_OR_LAB, ORGANIZATION
 - The default for FULL_NAME is the output of `git config --global user.name`. If it is 
   not set prompt the user for their full name. Also prompt the user for their DEFAULT_ROLE
   and allow 3 choices: Student, Faculty amd Staff Prompt the var DEPT_OR_LAB for Department/Lab
   and prompt ORGANIZATION for Oreganization and default to "Oregon State University" 
 
## sclabilty

 - implement parallelism and run one crawler/changer per sub directory

## Implementation Plan

### Language: Go

### Architecture Overview

The tool will be implemented as a command-line application in Go with the following components:

1. **main.go** - Entry point with CLI argument parsing
2. **config.go** - Configuration management
3. **filetypes.go** - File type to comment style mapping
4. **header.go** - Header generation logic
5. **crawler.go** - Parallel directory traversal
6. **processor.go** - File modification logic
7. **detector.go** - Existing header detection

### Detailed Component Design

#### 1. Configuration Management (config.go)
- **Config Structure**:
  ```go
  type Config struct {
      FullName     string `yaml:"FULL_NAME"`
      DefaultRole  string `yaml:"DEFAULT_ROLE"`  // Student, Faculty, or Staff
      DeptOrLab    string `yaml:"DEPT_OR_LAB"`
      Organization string `yaml:"ORGANIZATION"`
  }
  ```
- **Location**: `~/.config/licer.yml`
- **Creation Flow**:
  1. Check if config exists
  2. If not, get `git config --global user.name` for default name
  3. Prompt user for missing values interactively
  4. Save config in YAML format

#### 2. File Type Mapping (filetypes.go)
- **Comment Style Map**:
  ```go
  var commentStyles = map[string]CommentStyle{
      ".go":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
      ".py":   {Line: "#"},
      ".sh":   {Line: "#"},
      ".rb":   {Line: "#"},
      ".js":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
      ".ts":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
      ".java": {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
      ".c":    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
      ".cpp":  {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
      ".rs":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
      ".swift":{Line: "//", BlockStart: "/*", BlockEnd: "*/"},
      ".kt":   {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
      ".yaml": {Line: "#"},
      ".yml":  {Line: "#"},
      ".toml": {Line: "#"},
      ".sql":  {Line: "--", BlockStart: "/*", BlockEnd: "*/"},
      ".lua":  {Line: "--", BlockStart: "--[[", BlockEnd: "--]]"},
      ".r":    {Line: "#"},
      ".R":    {Line: "#"},
      ".m":    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
      ".vim":  {Line: "\""},
      ".el":   {Line: ";;"},
      ".lisp": {Line: ";;"},
      ".clj":  {Line: ";;"},
      ".hs":   {Line: "--", BlockStart: "{-", BlockEnd: "-}"},
      ".ml":   {Line: "(*", BlockEnd: "*)"},
      ".pas":  {Line: "//", BlockStart: "{", BlockEnd: "}"},
      ".pl":   {Line: "#"},
      ".php":  {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
      "":      {Line: "#"}, // No extension = shell script
  }
  ```
- **Excluded Extensions**: `.md`, `.json`, `.csv`, `.xml`, `.html`, `.txt`, `.log`, binary files

#### 3. Header Generation (header.go)

Based on role from config:
- **Student** → MIT License header
- **Faculty/Staff** → Apache 2.0 License header

**Header Templates**:

For Students (MIT):
```
# Copyright (c) {year} {FullName}
#
# SPDX-License-Identifier: MIT
# See LICENSE file for full license text.
```

For Faculty/Staff (Apache 2.0):
```
# Copyright {year} Oregon State University
#
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file for details.
# SPDX-License-Identifier: Apache-2.0
#
# Developed by: {FullName / DeptOrLab}
#               Oregon State University
#               {optional contact info}
```

#### 4. Header Detection (detector.go)

**Idempotency Check**:
- Search for "SPDX-License-Identifier" in first 20 lines of file
- If found, file already has a header (skip unless --force)
- When --force is used:
  1. Find the end of existing header (look for SPDX line or copyright block)
  2. Replace entire header block with new header

**Third-Party Copyright Detection**:
- Search for case-insensitive "Copyright" in first 3 lines of file
- If found without SPDX identifier, treat as third-party copyright
- Only overwrite third-party copyright headers when --force is used
- When --force is used with third-party headers:
  1. Detect the full copyright block (from Copyright line to end of license text)
  2. Replace entire third-party header with new header
  3. Preserve shebang if present

**Shebang Handling**:
- Detect shebang (`#!`) on first line
- If present, insert header starting from line 2
- Preserve the shebang line

#### 5. Parallel Crawler (crawler.go)

**Algorithm**:
```go
func ProcessDirectory(dir string) {
    entries := readDir(dir)
    
    // Create wait group for subdirectories
    var wg sync.WaitGroup
    
    // Process files in current directory
    for _, file := range entries {
        if isFile(file) && shouldProcess(file) {
            processFile(file)
        }
    }
    
    // Launch workers for subdirectories
    for _, subdir := range entries {
        if isDirectory(subdir) && !isGitDir(subdir) {
            wg.Add(1)
            go func(d string) {
                defer wg.Done()
                ProcessDirectory(d)
            }(subdir)
        }
    }
    
    wg.Wait()
}
```

**Key Points**:
- Skip `.git` directories
- Each subdirectory gets its own goroutine
- Recursive spawning creates a tree of workers
- Use sync.WaitGroup for synchronization
- Consider using a semaphore to limit max goroutines if needed

#### 6. File Processing (processor.go)

**Processing Flow**:
1. Read file
2. Detect if binary (skip if binary)
3. Check for existing header (unless --force)
4. Detect shebang
5. Generate appropriate header based on file type and config
6. Insert/replace header
7. Write file back
8. Log action if verbose

**Verbose Output Format**:
```
[ADD] src/main.go - Added Apache-2.0 header
[SKIP] src/util.go - Header already exists
[REPLACE] src/config.go - Replaced header (force mode)
[SKIP] data.json - Excluded file type
```

### CLI Arguments

```
licer [flags]

Flags:
  --git-folder string   Path to git repository (default: current directory)
  --force              Force replacement of existing headers
  --remove             Remove existing headers (requires SPDX-License-Identifier and organization/user match)
  --verbose            Verbose output (default: true)
  --help               Show help message
```

### Error Handling

- Exit with error if not in git repository
- Skip files that cannot be read/written (log warning)
- Handle missing config gracefully with prompts
- Validate config values (role must be Student/Faculty/Staff)

### Testing Strategy

1. Create test repository with various file types
2. Test idempotency (run twice, second run should change nothing)
3. Test --force flag (should replace existing headers)
4. Test shebang preservation
5. Test parallel processing with nested directories
6. Test excluded file types are skipped
7. Test config creation flow

### Performance Considerations

- Use buffered I/O for file operations
- Limit number of concurrent goroutines with semaphore if needed
- Process files in batches within each directory
- Consider file size limits (skip very large files)

## Header Removal Feature (--remove)

### Purpose
The `--remove` flag allows users to safely remove copyright headers that were previously added by Licer, providing a way to clean up or migrate licensing approaches.

### Safety Requirements
For security and legal compliance, headers can only be removed if they meet **ALL** of the following criteria:

1. **SPDX Identifier Present**: The header must contain the string "SPDX-License-Identifier:"
2. **Ownership Match**: The header must contain either:
   - The user's full name (from `FULL_NAME` in config), OR
   - The organization name (from `ORGANIZATION` in config)
3. **Valid Header Structure**: The header must be detectable using the same logic as the idempotency check

### Behavior
- **Mutually Exclusive**: `--remove` cannot be used with `--force`
- **Safe Operation**: Only removes headers that match the ownership criteria
- **Verbose Logging**: Shows which headers were removed vs. skipped
- **Preserves Shebangs**: Maintains shebang lines when removing headers

### Implementation Details

#### Header Validation Process:
1. Detect existing header using SPDX-License-Identifier
2. Extract header content between start and end lines
3. Check if header contains user's name OR organization name
4. Only remove if both SPDX and ownership criteria are met

#### Output Messages:
```
[REMOVE] src/main.go - Removed header (ownership match)
[SKIP] src/util.go - No header found
[SKIP] src/config.go - Header ownership mismatch (safety check)
[SKIP] vendor/lib.go - No SPDX identifier found
```

#### Ownership Matching Logic:
- **User Match**: Search for exact `FULL_NAME` string in header
- **Organization Match**: Search for exact `ORGANIZATION` string in header
- **Case Sensitive**: Exact string matching for security
- **Whitespace Tolerant**: Allow for different formatting/spacing

## LICENSE File Management

### Purpose
Licer automatically manages the root LICENSE file to ensure consistency with the headers it adds to source files.

### Behavior
When processing a repository, Licer checks for an existing `LICENSE` file in the repository root:

1. **No LICENSE file exists**: Creates appropriate LICENSE file based on user role
2. **LICENSE file exists with SPDX identifier**: Leaves it unchanged 
3. **LICENSE file exists without SPDX identifier**: 
   - Renames existing LICENSE to `LICENSE.orig`
   - Creates new LICENSE file matching the headers being added
   - If `LICENSE.orig` already exists, preserves both files unchanged

### SPDX Detection in LICENSE Files
- Searches entire LICENSE file for "SPDX-License-Identifier" (case-insensitive)
- If found, assumes LICENSE file is already managed/compatible with Licer
- If not found, treats as third-party or legacy license that should be preserved

### License File Templates
- **Students (MIT)**: Creates standard MIT license with user's name
- **Faculty/Staff (Apache 2.0)**: Creates Apache 2.0 license with Oregon State University copyright

### Safety Features
- **Preserves Original**: Never overwrites existing LICENSE.orig files
- **Non-Destructive**: Always backs up third-party licenses before replacement
- **Verbose Logging**: Reports LICENSE file operations clearly

### Output Messages
```
[LICENSE] Created LICENSE file (MIT)
[LICENSE] Renamed LICENSE to LICENSE.orig, created new LICENSE (Apache-2.0)
[LICENSE] LICENSE file already compatible (contains SPDX identifier)
[LICENSE] Skipped LICENSE management (LICENSE.orig exists)
```

## Git Pre-Commit Hook Integration

### Purpose
Automatically add license headers to newly created files before they are committed, ensuring all new code is properly licensed from the moment it enters the repository.

### CLI Options

#### `--hook` Flag (Hook Management)
Toggles the activation/deactivation of the Git pre-commit hook:
- `licer --hook` - Install or reinstall the pre-commit hook
- `licer --hook --remove` - Uninstall the pre-commit hook

#### `--pre-commit` Flag (Hook Execution Mode)  
Special mode used by the pre-commit hook to process only newly staged files:
- Only processes files that are being added to the repository (git status 'A')
- Uses `git diff --cached --name-status | grep '^A'` to find new files
- Skips files that are modified ('M') since they may already have headers
- Runs silently without interactive prompts
- Does not manage LICENSE files (to avoid conflicts during commit)

### Behavioral Rules

#### Interactive Hook Activation
When `licer` is run with no options in a repository:
1. Check if pre-commit hook is installed
2. If not installed, prompt user: "Install pre-commit hook to automatically license new files? (y/N): "
3. If user responds 'y' or 'Y', install the hook
4. Continue with normal repository processing

#### Unattended Mode
When `licer --git-folder <path>` is used:
- **Never prompt** for hook installation
- **Never install hooks** automatically
- Process the repository silently
- Used for batch processing or CI/CD scenarios

### Hook Implementation

#### Hook Script Content
The pre-commit hook script (`.git/hooks/pre-commit`):
```bash
#!/bin/bash

# Get the directory where licer binary is located
LICER_PATH="$(which licer)"
if [ -z "$LICER_PATH" ]; then
    # Try to find licer in common locations
    for path in "./licer" "../licer" "$(git rev-parse --show-toplevel)/licer"; do
        if [ -x "$path" ]; then
            LICER_PATH="$path"
            break
        fi
    done
fi

if [ -z "$LICER_PATH" ]; then
    echo "Warning: licer not found, skipping header check"
    exit 0
fi

# Run licer in pre-commit mode
"$LICER_PATH" --pre-commit --verbose=false

exit 0
```

#### Hook Installation Process
1. Check if `.git/hooks/pre-commit` exists
2. If exists, back up to `.git/hooks/pre-commit.backup` 
3. Create new pre-commit hook with licer integration
4. Make hook executable (`chmod +x`)
5. Report installation status

#### Hook Detection
Check for hook installation by:
1. Verify `.git/hooks/pre-commit` exists and is executable
2. Check if content contains "licer --pre-commit" 
3. If both conditions met, hook is considered installed

### File Processing Logic

#### Staged File Detection
```go
func getStagedNewFiles() ([]string, error) {
    cmd := exec.Command("git", "diff", "--cached", "--name-status")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    var newFiles []string
    lines := strings.Split(string(output), "\n")
    
    for _, line := range lines {
        if strings.HasPrefix(line, "A\t") {
            filename := strings.TrimPrefix(line, "A\t")
            newFiles = append(newFiles, filename)
        }
    }
    
    return newFiles, nil
}
```

#### Pre-commit Processing Flow
1. Get list of staged files with 'A' (added) status
2. For each new file:
   - Check if file type is supported
   - Check if file already has header (skip if yes)
   - Add appropriate header based on user config
   - Stage the modified file (`git add <file>`)
3. Report results silently (only errors to stderr)
4. Exit with code 0 for success, 1 for errors

### Command Examples

```bash
# Install pre-commit hook
licer --hook

# Uninstall pre-commit hook  
licer --hook --remove

# Manual pre-commit processing (used by hook)
licer --pre-commit

# Normal processing (may prompt for hook installation)
licer

# Batch mode (no hook prompts)
licer --git-folder /path/to/repo
```

### Safety & Error Handling

#### Hook Installation Safety
- **Backup existing hooks** before replacement
- **Validate Git repository** before hook installation
- **Check write permissions** in .git/hooks directory
- **Verify licer binary exists** and is executable

#### Pre-commit Mode Safety
- **Non-interactive**: No prompts during commit process
- **Fast execution**: Only process newly added files
- **Minimal output**: Silent unless errors occur
- **Graceful failure**: Don't block commits on minor errors

#### Error Scenarios
- **Licer not found**: Hook warns but allows commit to proceed
- **File processing error**: Log warning, continue with other files
- **Permission error**: Report to stderr, exit with code 1

### Future Enhancements

1. Add --dry-run flag to preview changes
2. Support custom license types via config
3. Add --exclude flag for custom exclusion patterns  
4. Support for multi-line copyright (contributors)
5. Progress bar for large repositories
6. License detection based on LICENSE file in repo
7. Support for different header styles per subdirectory
8. Add --remove-all flag for admin override (with confirmation prompt)
9. Add --remove-match="pattern" for custom removal criteria
10. Add --skip-license flag to disable LICENSE file management
11. Add hook configuration options in licer.yml
12. Support for other Git hooks (post-commit, pre-push)
13. Integration with popular pre-commit frameworks 


