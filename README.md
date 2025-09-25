# Licer 📄⚖️

**License Header Management Tool for Oregon State University**

Licer is a powerful command-line tool that automatically adds, manages, and removes copyright headers in source code files across entire Git repositories. Built specifically for Oregon State University's intellectual property requirements, it ensures consistent licensing across all your projects.

<img width="814" height="347" alt="image" src="https://github.com/user-attachments/assets/93dd2d0f-986a-41c1-8993-c3ef47d7b662" />



## ✨ Features

### 🔧 **Core Functionality**
- **Recursive Processing**: Crawls entire Git repositories automatically
- **Parallel Processing**: Uses concurrent workers for fast processing of large codebases
- **Git Integration**: Only works in Git repositories for safety
- **Role-Based Licensing**: Different headers for Students vs Faculty/Staff
- **Idempotency**: Safe to run multiple times without duplication

### 📝 **License Management**
- **Students**: MIT License headers with personal copyright
- **Faculty/Staff**: Apache 2.0 headers with Oregon State University copyright
- **SPDX Compliance**: All headers include SPDX-License-Identifier tags
- **Policy Compliant**: Implements OSU Policy 06-200 requirements

### 🛡️ **Safety Features**
- **Third-Party Protection**: Detects and protects third-party copyrights
- **Force Override**: `--force` flag for intentional third-party replacement  
- **Ownership Verification**: `--remove` only removes headers you own
- **Shebang Preservation**: Maintains script shebang lines
- **Backup Creation**: LICENSE files backed up as LICENSE.orig

### 🌐 **File Type Support**
Supports 25+ programming languages and file types:

| Language | Extensions | Comment Style |
|----------|------------|---------------|
| **Web Development** | `.html`, `.htm`, `.css`, `.scss`, `.sass`, `.less` | `<!-- -->`, `/* */` |
| **JavaScript** | `.js`, `.mjs`, `.cjs`, `.ts`, `.tsx`, `.jsx` | `//`, `/* */` |
| **Python** | `.py` | `#` |
| **Go** | `.go` | `//`, `/* */` |
| **C/C++** | `.c`, `.cpp`, `.cc`, `.cxx`, `.h`, `.hpp` | `//`, `/* */` |
| **Java** | `.java` | `//`, `/* */` |
| **Rust** | `.rs` | `//`, `/* */` |
| **R** | `.r`, `.R`, `.rmd`, `.Rmd` | `#`, `<!-- -->` |
| **Shell** | `.sh`, No extension | `#` |
| **Ruby** | `.rb` | `#` |
| **Configuration** | `.yaml`, `.yml`, `.toml`, `.ini`, `.cfg`, `.conf` | `#` |
| **SQL** | `.sql` | `--`, `/* */` |
| **And many more...** | See filetypes.go | Various |

## 🚀 Installation

### Prerequisites
- Go 1.19 or later
- Git repository (Licer only works in Git repos)

### Build from Source
```bash
git clone https://github.com/dirkpetersen/licer.git
cd licer
go build -o licer ./src
```

## 📖 Usage

### Basic Commands

```bash
# Process current Git repository
licer

# Process specific repository
licer --git-folder /path/to/repo

# Replace existing headers
licer --force

# Remove headers (safe mode - only removes headers you own)
licer --remove

# Quiet mode
licer --verbose=false

# Show help
licer --help
```

### First Run Setup
On first run, Licer will prompt you to create a configuration file:

```
Full Name (default: John Doe): 
Role (1=Student, 2=Faculty, 3=Staff): 2
Department/Lab: Computer Science Department
Organization (default: Oregon State University): 
```

Configuration is saved to `~/.config/licer.yml`

## 🎯 Examples

### Student Project (MIT License)
```python
# Copyright (c) 2025 Jane Smith
#
# SPDX-License-Identifier: MIT
# See LICENSE file for full license text.

def hello_world():
    print("Hello, World!")
```

### Faculty/Staff Project (Apache 2.0)
```python
# Copyright 2025 Oregon State University
#
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file for details.
# SPDX-License-Identifier: Apache-2.0
#
# Developed by: Dr. John Doe
#               Computer Science Department

def research_function():
    return "Important research code"
```

### CSS Files (Block Comments)
```css
/*
 * Copyright 2025 Oregon State University
 *
 * Licensed under the Apache License, Version 2.0.
 * See the LICENSE file for details.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Developed by: Web Development Team
 *               UIT/ARCS
 */

body {
    font-family: Arial, sans-serif;
}
```

### Shell Scripts (Shebang Preserved)
```bash
#!/bin/bash

# Copyright 2025 Oregon State University
#
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file for details.
# SPDX-License-Identifier: Apache-2.0
#
# Developed by: System Administration
#               UIT/ARCS

echo "Deployment script"
```

## 🔒 Security & Safety

### Third-Party Copyright Protection
Licer detects third-party copyrights and protects them:

```bash
# This will be SKIPPED without --force
# Copyright (c) 2020 Some Other Company

# This requires --force to overwrite
licer --force  # Only use when you have permission!
```

### Safe Header Removal
The `--remove` flag only removes headers that contain:
- Your full name (from config), OR
- Your organization name (from config)
- AND contain "SPDX-License-Identifier"

```bash
# Only removes headers you own
licer --remove

# Will NOT remove third-party headers for safety
```

### LICENSE File Management
Licer automatically manages the root LICENSE file:

1. **No LICENSE**: Creates appropriate LICENSE file
2. **LICENSE with SPDX**: Leaves unchanged
3. **Third-party LICENSE**: Renames to LICENSE.orig, creates new LICENSE
4. **LICENSE.orig exists**: Preserves both files unchanged

## 📋 Command Reference

| Flag | Description |
|------|-------------|
| `--git-folder` | Path to Git repository (default: current directory) |
| `--force` | Force replacement of existing headers (including third-party) |
| `--remove` | Remove headers safely (only removes headers you own) |
| `--verbose` | Verbose output (default: true) |
| `--help` | Show help message |

## 🔍 Verbose Output

Licer provides detailed logging of all operations:

```
[ADD] src/main.py - Added Apache-2.0 header
[REPLACE] src/util.py - Replaced header (force mode)
[REMOVE] src/old.py - Removed header (ownership match)
[SKIP] src/third_party.py - Third-party copyright found (use --force to overwrite)
[SKIP] README.md - Excluded file type
[LICENSE] Renamed LICENSE to LICENSE.orig, created new LICENSE (Apache-2.0)

=== Processing Summary ===
Files processed: 156
Files modified:  89
Files skipped:   67
=========================
```

## 🏛️ Oregon State University Policy Compliance

Licer implements [OSU Policy 06-200 Intellectual Property](https://policy.oregonstate.edu/06-200) requirements:

- **Students**: Personal ownership with MIT licensing (Policy 5.1.3)
- **Faculty/Staff**: University ownership with Apache 2.0 licensing (Policy 4.8, 4.11, 4.18)
- **Sponsored Research**: University ownership with appropriate attribution
- **SPDX Compliance**: Industry-standard license identification

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🐛 Issues & Support

- Report bugs: [GitHub Issues](https://github.com/your-org/licer/issues)
- Documentation: See [CLAUDE.md](CLAUDE.md) for implementation details
- Contact: [Your Contact Information]

## 🙏 Acknowledgments

- Oregon State University Office for Commercialization and Corporate Development (OCCD)
- OSU Policy 06-200 Intellectual Property guidelines
- Open source community for SPDX standards

---

**Made with ❤️ for Oregon State University**

*Ensuring legal compliance and protecting intellectual property rights across all university software projects.*
