package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	FullName     string `yaml:"FULL_NAME"`
	DefaultRole  string `yaml:"DEFAULT_ROLE"`
	DeptOrLab    string `yaml:"DEPT_OR_LAB"`
	Organization string `yaml:"ORGANIZATION"`
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, ".config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	
	return filepath.Join(configDir, "licer.yml"), nil
}

func LoadOrCreateConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	
	// Try to load existing config
	if _, err := os.Stat(configPath); err == nil {
		return loadConfig(configPath)
	}
	
	// Create new config
	config, err := createConfig()
	if err != nil {
		return nil, err
	}
	
	// Save config
	if err := saveConfig(config, configPath); err != nil {
		return nil, err
	}
	
	return config, nil
}

func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Validate required fields
	if config.FullName == "" || config.DefaultRole == "" || 
	   config.DeptOrLab == "" || config.Organization == "" {
		return nil, fmt.Errorf("config file is incomplete, please delete it and run again to recreate")
	}
	
	// Validate role
	if config.DefaultRole != "Student" && config.DefaultRole != "Faculty" && config.DefaultRole != "Staff" {
		return nil, fmt.Errorf("invalid role '%s', must be Student, Faculty, or Staff", config.DefaultRole)
	}
	
	return &config, nil
}

func createConfig() (*Config, error) {
	config := &Config{}
	reader := bufio.NewReader(os.Stdin)
	
	// Get full name with git fallback
	gitName := getGitUserName()
	if gitName != "" {
		fmt.Printf("Full Name (default: %s): ", gitName)
	} else {
		fmt.Print("Full Name: ")
	}
	
	nameInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	
	nameInput = strings.TrimSpace(nameInput)
	if nameInput == "" && gitName != "" {
		config.FullName = gitName
	} else if nameInput != "" {
		config.FullName = nameInput
	} else {
		return nil, fmt.Errorf("full name is required")
	}
	
	// Get role
	for {
		fmt.Print("Role (1=Student, 2=Faculty, 3=Staff): ")
		roleInput, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}
		
		roleInput = strings.TrimSpace(roleInput)
		switch roleInput {
		case "1":
			config.DefaultRole = "Student"
		case "2":
			config.DefaultRole = "Faculty"
		case "3":
			config.DefaultRole = "Staff"
		default:
			fmt.Println("Please enter 1, 2, or 3")
			continue
		}
		break
	}
	
	// Get department/lab
	fmt.Print("Department/Lab: ")
	deptInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	config.DeptOrLab = strings.TrimSpace(deptInput)
	if config.DeptOrLab == "" {
		return nil, fmt.Errorf("department/lab is required")
	}
	
	// Get organization
	fmt.Print("Organization (default: Oregon State University): ")
	orgInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	
	orgInput = strings.TrimSpace(orgInput)
	if orgInput == "" {
		config.Organization = "Oregon State University"
	} else {
		config.Organization = orgInput
	}
	
	return config, nil
}

func saveConfig(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	fmt.Printf("Configuration saved to %s\n", configPath)
	return nil
}

func getGitUserName() string {
	cmd := exec.Command("git", "config", "--global", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}