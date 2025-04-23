# Golang Markdown Editor

A minimalist markdown editor built with Go and Fyne, optimized for Arch Linux.

## Features

- **Folder-based Workspace**: Open directories and manage markdown files
- **Live Preview**: Real-time markdown rendering
- **File Operations**:
  - Create new markdown files (`.md` extension enforced)
  - Edit and save existing files
  - Delete files with confirmation

## Installation

### Requirements
- **Go 1.21+** (build requirement)

### Build & Run
1. Clone repository:
   ```bash
   git clone https://github.com/sokolawesome/golang-markdown-editor.git
   cd golang-markdown-editor
   ```

2. Build executable:
   ```bash
   mkdir -p bin
   go build -o bin/markdown-editor ./cmd/markdown-editor
   ```

3. Launch application:
   ```bash
   ./bin/markdown-editor
   ```

## First Run Setup

1. **Select Workspace**:
   Choose your default markdown folder when prompted

2. **Configuration File**:
   Automatically created at:
   `~/.config/markdown-editor/config.json`

## Interface Overview

- **Left Panel**
  File tree showing all `.md` files in workspace
- **Right Panels**
  - **Top**: Raw markdown editor
  - **Bottom**: Formatted preview

## Platform Compatibility

**Primary Supported OS**:
- Arch Linux (tested on latest kernel)

**Untested But Possible**:
- Other Linux distributions (may require additional setup)
- Windows/macOS (not officially supported)

## Configuration

To modify default workspace, edit:
`~/.config/markdown-editor/config.json`

```json
{
  "default_folder": "/path/to/your/notes"
}
```

> The application will automatically create this file and directory structure on first run
