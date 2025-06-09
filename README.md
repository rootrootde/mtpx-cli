# mtpx-cli

A command-line interface for interacting with MTP (Media Transfer Protocol) devices, such as Android phones and tablets.

## Features

- List files and directories on MTP devices
- Download files from MTP devices
- Upload files to MTP devices
- Delete files from MTP devices
- Check file existence and get file information
- Display device and storage information
- JSON-formatted output for easy integration with other tools
- Progress tracking for file transfers

## Installation

### Prerequisites

- Go 1.24.3 or later
- pkg-config
- libusb

On macOS, install dependencies with Homebrew:
```bash
brew install pkg-config libusb
```

### Building from source

```bash
./build.sh
```

Or manually:
```bash
go build -o mtpx-cli main.go
```

## Usage

```
mtpx-cli <command> [arguments]
```

### Commands

#### List files
List files and directories at a remote path:
```bash
./mtpx-cli list <remote_path>
```

Example:
```bash
./mtpx-cli list /DCIM/Camera
```

#### Download files
Download a file from the device to a local directory:
```bash
./mtpx-cli download <remote_file> <local_directory>
```

Example:
```bash
./mtpx-cli download /DCIM/Camera/IMG_001.jpg ./downloads/
```

#### Upload files
Upload a local file to a directory on the device:
```bash
./mtpx-cli upload <local_file> <remote_directory>
```

Example:
```bash
./mtpx-cli upload ./photo.jpg /DCIM/Camera/
```

#### Delete files
Delete one or more files from the device:
```bash
./mtpx-cli delete <remote_path> [<remote_path2> ...]
```

Example:
```bash
./mtpx-cli delete /DCIM/Camera/IMG_001.jpg /DCIM/Camera/IMG_002.jpg
```

#### Check file existence
Check if a file exists and display its size:
```bash
./mtpx-cli stat <remote_path>
```

Example:
```bash
./mtpx-cli stat /DCIM/Camera/IMG_001.jpg
```

#### Device information
Display basic device information:
```bash
./mtpx-cli device-info
```

#### Storage information
Display storage-related information:
```bash
./mtpx-cli storage-info
```

## Output Format

All commands output JSON-formatted data for easy parsing and integration with other tools. Each operation includes a completion sentinel (e.g., `MTPX_LIST_DONE`, `MTPX_DOWNLOAD_DONE`) to indicate when the operation has finished.

### Progress Updates

File transfers (upload/download) emit progress updates in JSON format:
```json
{
  "file": "IMG_001.jpg",
  "progress": 45.5
}
```

### Transfer Summary

Upon completion, transfers output source and target paths:
```json
{
  "source": "/DCIM/Camera/IMG_001.jpg",
  "target": "./downloads/IMG_001.jpg"
}
```

## Architecture

The codebase is organized with:
- `CLI` struct that encapsulates device and storage management
- Separate handler functions for each command
- `ProgressHandler` struct for consistent progress reporting
- JSON output helper functions for standardized formatting
- Comprehensive error handling with contextual information

## Dependencies

- [github.com/ganeshrvel/go-mtpx](https://github.com/ganeshrvel/go-mtpx) - MTP operations library
- [github.com/ganeshrvel/go-mtpfs](https://github.com/ganeshrvel/go-mtpfs) - MTP filesystem implementation

## License

This project is open source. Please check the repository for license information.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.