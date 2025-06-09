# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Go-based CLI tool for interacting with MTP (Media Transfer Protocol) devices. It provides functionality to list, download, upload, delete files, and retrieve device information from MTP-connected devices like Android phones.

## Architecture

The codebase is organized as follows:
- `main.go` - Main implementation with clean, modular structure:
  - `CLI` struct encapsulates device and storage
  - Separate handler functions for each command
  - `ProgressHandler` struct eliminates code duplication
  - JSON output helper functions for consistent formatting
  - Better error handling with context
- Uses the `github.com/ganeshrvel/go-mtpx` library for MTP operations
- Uses the `github.com/ganeshrvel/go-mtpfs/mtp` library for device types
- Outputs JSON-formatted progress and results for machine parsing
- Uses sentinel values (e.g., `MTPX_LIST_DONE`) to indicate operation completion

## Common Commands

### Building the project
```bash
./build.sh
```
Or manually:
```bash
go build -o mtpx-cli main.go
```

### Running the CLI
```bash
./mtpx-cli <command> [arguments]
```

Available commands:
- `list <remote_path>` - List files at remote path
- `download <remote> <local_dir>` - Download a file into target directory
- `upload <local_file> <remote_dir>` - Upload a file into remote directory
- `delete <remote_path> [...]` - Delete one or more files by remote path
- `stat <remote_path>` - Check if a file exists and print its size
- `device-info` - Show basic device information
- `storage-info` - Show storage-related information

### Managing dependencies
```bash
go mod tidy
go mod download
```

## Key Implementation Details

- The tool automatically connects to the first available MTP storage device
- All output is JSON-formatted for easy parsing by other tools
- Progress updates are emitted during upload/download operations
- Each command prints a completion sentinel (e.g., `MTPX_DOWNLOAD_DONE`)
- Error handling uses log.Fatal() for immediate termination with error messages