package main

import (
	"fmt"
	"time"
)

func GenerateHeader(config *Config) string {
	year := time.Now().Year()
	
	switch config.DefaultRole {
	case "Student":
		return generateStudentHeader(config, year)
	case "Faculty", "Staff":
		return generateFacultyStaffHeader(config, year)
	default:
		// Default to student if role is unclear
		return generateStudentHeader(config, year)
	}
}

func generateStudentHeader(config *Config, year int) string {
	return fmt.Sprintf(`Copyright (c) %d %s

SPDX-License-Identifier: MIT
See LICENSE file for full license text.`, year, config.FullName)
}

func generateFacultyStaffHeader(config *Config, year int) string {
	return fmt.Sprintf(`Copyright %d Oregon State University

Licensed under the Apache License, Version 2.0.
See the LICENSE file for details.
SPDX-License-Identifier: Apache-2.0

Developed by: %s
              %s`, year, config.FullName, config.DeptOrLab)
}

func GetHeaderTemplate(config *Config) HeaderTemplate {
	switch config.DefaultRole {
	case "Student":
		return HeaderTemplate{
			LicenseType: "MIT",
			CopyrightOwner: config.FullName,
		}
	case "Faculty", "Staff":
		return HeaderTemplate{
			LicenseType: "Apache-2.0",
			CopyrightOwner: "Oregon State University",
		}
	default:
		return HeaderTemplate{
			LicenseType: "MIT",
			CopyrightOwner: config.FullName,
		}
	}
}

type HeaderTemplate struct {
	LicenseType     string
	CopyrightOwner  string
}