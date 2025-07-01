# fjrd

A macOS configuration management tool that applies system settings via TOML configuration files.

## Features

- TOML configuration file
- `defaults` write command generation
- Install applications with Homebrew

## Installation

```bash
go install github.com/RATIU5/fjrd/cmd/fjrd@latest
```

Or build from source:

```bash
git clone https://github.com/RATIU5/fjrd
cd fjrd
go build ./cmd/fjrd
```

## Usage

### Basic Usage

```bash
# Apply local config
fjrd config.toml

# Apply config from GitHub repo
fjrd owner/repo

# Apply config from GitHub file
fjrd https://github.com/owner/repo/blob/main/fjrd.toml

# Apply config from HTTPS URL
fjrd https://example.com/config.toml
```

### Logging and Debugging

```bash
# Verbose logging (debug level)
fjrd -verbose config.toml

# Set specific log level
fjrd -log-level=info config.toml

# JSON formatted logs
fjrd -log-format=json config.toml

# Quiet mode (errors only)
fjrd -quiet config.toml

# Custom timeout
fjrd -timeout=60s config.toml
```

### Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-log-level` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `-log-format` | `text` | Log format: `text`, `json` |
| `-verbose` | `false` | Enable debug logging (same as `-log-level=debug`) |
| `-quiet` | `false` | Suppress non-error output |
| `-timeout` | `30s` | Operation timeout (e.g., `60s`, `2m`) |
| `-help` | `false` | Show help message |

## Configuration

### Basic Example

```toml
version = 1

[macos.dock]
autohide = false

[macos.defaultsRaw]
"com.apple.dock.workspaces-auto-swoosh" = { value = 0, type = "int" }
```

### Dock Configuration Options

| Option | Type | Description | Values |
|--------|------|-------------|---------|
| `autohide` | bool | Enable dock autohide | `true`, `false` |
| `orientation` | string | Dock position | `"left"`, `"bottom"`, `"right"` |
| `tilesize` | int | Icon size in pixels | Any positive integer |
| `autohide-time-modifier` | float | Autohide animation speed | 0.0 - 2.0 |
| `autohide-delay` | float | Autohide delay in seconds | 0.0 - 2.0 |
| `show-recents` | bool | Show recent apps | `true`, `false` |
| `mineffect` | string | Minimize effect | `"genie"`, `"scale"`, `"suck"` |
| `static-only` | bool | Show only active apps | `true`, `false` |
| `scroll-to-open` | bool | Scroll to open apps | `true`, `false` |

### Raw Defaults

Use `defaultsRaw` to apply any macOS defaults setting:

```toml
[macos.defaultsRaw]
"com.apple.dock.setting-name" = { value = "setting-value", type = "string" }
```

Supported types:
- `"string"` - Text values
- `"bool"` - Boolean values (`true`/`false`)
- `"int"` - Integer numbers
- `"float"` - Decimal numbers

### Resetting to System Defaults

You can reset any setting to its system default using three methods:

```toml
[macos.defaultsRaw]
# Method 1: Use "default" string
"com.apple.dock.workspaces-auto-swoosh" = { value = "default", type = "int" }

# Method 2: Use explicit reset flag 
"com.apple.dock.expose-animation-duration" = { value = 0.1, type = "float", reset = true }

# Method 3: Omit the setting entirely (no change)
# "com.apple.dock.launchanim" = { value = false, type = "bool" }
```

When a setting is reset:
- The application runs `defaults delete domain.key`
- The setting returns to its system/application default
- Useful for cleaning up previous customizations

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

[See the LICENSE](LICENSE.md)
