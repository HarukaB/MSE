# MSE (Muse Save Editor)
[中文](./README_CN.md)
A command-line tool written in Go for serializing and deserializing Muse Dash save files (`.sav`). It allows converting binary save files to readable JSON format for editing and convert them back to the original binary format.

## Features

- **SAV to JSON**: Convert Muse Dash binary save files to human-readable JSON.
- **JSON to SAV**: Convert modified JSON files back to the binary format used by the game.

## Installation

### Download from Releases (Recommended)

Download the pre-built binaries for Windows, macOS, and Linux directly from the [Releases](../../releases) page. Run the executable directly.

### Build from Source

To build from source, ensure [Go](https://golang.org/dl/) is installed (version 1.21 or later).

1. Clone the repository or download the source code.
2. Build the executable:
   ```bash
   go build -o mse main.go
   ```

## Usage

The tool provides two main commands: `to-json` and `to-sav`.

### Convert SAV to JSON

To view or edit a save file, convert it to JSON:

```bash
./mse to-json input.sav output.json
```

### Convert JSON to SAV

After making changes to the JSON file, convert it back to the binary `.sav` format:

```bash
./mse to-sav input.json output.sav
```

## Project Structure

- `main.go`: Entry point for the CLI application.
- `converter/`: High-level logic for file conversion.
- `odin/`: Core implementation of the Odin Serializer format.
  - `reader.go`: Binary data reader for parsing `.sav` files.
  - `writer.go`: Binary data writer for generating `.sav` files.
  - `node.go`: Data structures representing the serialization tree.
  - `entry.go`: Constants and types for the Odin format.

## Disclaimer

This tool is for educational and modding purposes. Always back up save files before modification. The authors are not responsible for any data loss or account issues resulting from the use of this tool.
