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
	"testing"
)

func testConfig() *Config {
	return &Config{
		FullName:     "Test User",
		DefaultRole:  "Staff",
		DeptOrLab:    "Test Lab",
		Organization: "Oregon State University",
	}
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestFormatHeaderLineComments(t *testing.T) {
	style := commentStyles[".go"]
	out := FormatHeader("Copyright 2025 Test\n\nSPDX-License-Identifier: MIT", style)

	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "//") {
			t.Errorf("line does not start with //: %q", line)
		}
	}
}

func TestFormatHeaderCSSBlock(t *testing.T) {
	style := commentStyles[".css"]
	out := FormatHeader("Copyright 2025 Test", style)

	lines := strings.Split(out, "\n")
	if lines[0] != "/*" || lines[len(lines)-1] != " */" {
		t.Errorf("CSS header is not a /* ... */ block:\n%s", out)
	}
}

func TestFormatHeaderHTMLIsValid(t *testing.T) {
	style := commentStyles[".html"]
	out := FormatHeader("Copyright 2025 Test\n\nSPDX-License-Identifier: MIT", style)

	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "<!--") || !strings.HasSuffix(line, "-->") {
			t.Errorf("HTML header line is not a closed comment: %q", line)
		}
	}
}

func TestFormatHeaderOCamlIsValid(t *testing.T) {
	style := commentStyles[".ml"]
	out := FormatHeader("Copyright 2025 Test", style)

	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "(*") || !strings.HasSuffix(line, "*)") {
			t.Errorf("OCaml header line is not a closed comment: %q", line)
		}
	}
}

func TestLicenseFilesAreExcluded(t *testing.T) {
	for _, name := range []string{"LICENSE", "LICENSE.orig", "COPYING", "NOTICE", "license"} {
		path := writeTempFile(t, name, "Apache License\nVersion 2.0, January 2004\n")
		if ShouldProcessFile(path) {
			t.Errorf("%s should be excluded from processing", name)
		}
	}
}

func TestAddHeaderIsIdempotent(t *testing.T) {
	path := writeTempFile(t, "example.py", "def main():\n    pass\n")
	config := testConfig()

	result := ProcessFile(path, config, false, false, false)
	if result.Action != "ADD" || !result.Modified {
		t.Fatalf("expected ADD, got %s (%s)", result.Action, result.Reason)
	}

	result = ProcessFile(path, config, false, false, false)
	if result.Action != "SKIP" || result.Modified {
		t.Fatalf("second run should SKIP, got %s (%s)", result.Action, result.Reason)
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "SPDX-License-Identifier: Apache-2.0") {
		t.Error("header missing SPDX identifier")
	}
	if !strings.Contains(string(content), "def main():") {
		t.Error("original code was lost")
	}
}

func TestForceReplaceIsStable(t *testing.T) {
	path := writeTempFile(t, "example.py", "def main():\n    pass\n")
	config := testConfig()

	ProcessFile(path, config, false, false, false)
	ProcessFile(path, config, true, false, false)
	first, _ := os.ReadFile(path)
	ProcessFile(path, config, true, false, false)
	second, _ := os.ReadFile(path)

	if string(first) != string(second) {
		t.Errorf("repeated --force runs changed the file:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
	if !strings.Contains(string(second), "def main():") {
		t.Error("original code was lost during force replace")
	}
}

func TestShebangIsPreserved(t *testing.T) {
	path := writeTempFile(t, "deploy.sh", "#!/bin/bash\necho hello\n")
	config := testConfig()

	result := ProcessFile(path, config, false, false, false)
	if !result.Modified {
		t.Fatalf("expected file to be modified, got %s (%s)", result.Action, result.Reason)
	}

	content, _ := os.ReadFile(path)
	lines := strings.Split(string(content), "\n")
	if lines[0] != "#!/bin/bash" {
		t.Errorf("shebang not preserved as first line, got %q", lines[0])
	}
	if !strings.Contains(string(content), "echo hello") {
		t.Error("original code was lost")
	}

	// Force replace must also keep the shebang
	ProcessFile(path, config, true, false, false)
	content, _ = os.ReadFile(path)
	if !strings.HasPrefix(string(content), "#!/bin/bash") {
		t.Error("shebang lost after force replace")
	}
}

func TestThirdPartyCopyrightIsProtected(t *testing.T) {
	source := "// Copyright (c) 2020 Other Corp\n\nuse std::io;\n\nfn main() {}\n"
	path := writeTempFile(t, "lib.rs", source)
	config := testConfig()

	result := ProcessFile(path, config, false, false, false)
	if result.Action != "SKIP" || result.Modified {
		t.Fatalf("third-party copyright should be skipped without --force, got %s (%s)", result.Action, result.Reason)
	}

	// With --force the header is replaced but code must survive
	result = ProcessFile(path, config, true, false, false)
	if !result.Modified {
		t.Fatalf("expected --force to replace third-party header, got %s (%s)", result.Action, result.Reason)
	}

	content, _ := os.ReadFile(path)
	if strings.Contains(string(content), "Other Corp") {
		t.Error("third-party copyright not replaced under --force")
	}
	if !strings.Contains(string(content), "use std::io;") || !strings.Contains(string(content), "fn main() {}") {
		t.Errorf("code lines were lost during third-party replacement:\n%s", content)
	}
}

func TestCodeStartingWithCIsNotAComment(t *testing.T) {
	if isCommentLine("Config = load()") {
		t.Error("code starting with 'C' misdetected as comment")
	}
	if isCommentLine(`"""Module docstring."""`) {
		t.Error("Python docstring misdetected as comment")
	}
	if !isCommentLine("C Fortran comment") {
		t.Error("Fortran comment not detected")
	}
	if !isCommentLine("# shell comment") || !isCommentLine("// go comment") {
		t.Error("standard comments not detected")
	}
}

func TestRemoveHeaderWithOwnershipMatch(t *testing.T) {
	path := writeTempFile(t, "example.py", "def main():\n    pass\n")
	config := testConfig()

	ProcessFile(path, config, false, false, false)
	result := ProcessFile(path, config, false, true, false)
	if result.Action != "REMOVE" || !result.Modified {
		t.Fatalf("expected REMOVE, got %s (%s)", result.Action, result.Reason)
	}

	content, _ := os.ReadFile(path)
	if strings.Contains(string(content), "SPDX-License-Identifier") {
		t.Error("header not removed")
	}
	if !strings.Contains(string(content), "def main():") {
		t.Error("original code was lost during removal")
	}
}

func TestRemoveHeaderOwnershipMismatch(t *testing.T) {
	source := "# Copyright (c) 2025 Someone Else\n#\n# SPDX-License-Identifier: MIT\n\ndef main():\n    pass\n"
	path := writeTempFile(t, "example.py", source)

	result := ProcessFile(path, testConfig(), false, true, false)
	if result.Action != "SKIP" || result.Modified {
		t.Fatalf("foreign header should not be removed, got %s (%s)", result.Action, result.Reason)
	}

	content, _ := os.ReadFile(path)
	if string(content) != source {
		t.Error("file was modified despite ownership mismatch")
	}
}

func TestHookInstallDetection(t *testing.T) {
	repoRoot := t.TempDir()
	hooksDir := filepath.Join(repoRoot, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	if isHookInstalled(repoRoot) {
		t.Error("hook reported installed before installation")
	}

	if err := installPreCommitHook(repoRoot, false); err != nil {
		t.Fatalf("failed to install hook: %v", err)
	}
	if !isHookInstalled(repoRoot) {
		t.Error("hook not detected after installation")
	}

	if err := uninstallPreCommitHook(repoRoot, false); err != nil {
		t.Fatalf("failed to uninstall hook: %v", err)
	}
	if isHookInstalled(repoRoot) {
		t.Error("hook still detected after uninstallation")
	}
}
