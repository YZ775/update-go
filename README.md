# update-go

A CLI tool to list and install Go versions interactively.

## Features

- Fetch available Go versions from [go.dev](https://go.dev/dl/)
- Interactive version selection with search
- Automatic download with progress bar
- SHA256 checksum verification
- Seamless installation with backup of existing Go

## Installation

```bash
go install github.com/YZ775/update-go@latest
```

Or build from source:

```bash
git clone https://github.com/YZ775/update-go.git
cd update-go
go build -o update-go
```

## Usage

### Interactive Mode (default)

```bash
update-go
```

This will:
1. Fetch available Go versions
2. Display an interactive selection menu
3. Ask for confirmation before installing
4. Download and install the selected version

### List Versions Only

```bash
update-go -n
```

### Show All Versions (including unstable)

```bash
update-go -a
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--all` | `-a` | Show all versions (default: stable only) |
| `--no-interact` | `-n` | Disable interactive mode, only show list |

## Requirements

- Go 1.22 or later
- macOS or Linux (Windows is not currently supported)
- `sudo` access may be required for installation to `/usr/local/go`

## How It Works

1. Fetches version information from the official Go download API
2. Filters and displays stable versions (or all with `-a`)
3. Downloads the selected version to a temporary directory
4. Verifies the SHA256 checksum
5. Backs up existing Go installation (if present)
6. Extracts and installs to `/usr/local/go`

## License

MIT
