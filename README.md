# fjrd

A declarative macOS configuration management tool that applies system settings through simple TOML files. Configure your entire macOS environment with version-controlled settings that can be easily shared, replicated, and applied across multiple machines.

## What is fjrd?

**fjrd** (pronounced "ford") is a command-line tool that bridges the gap between macOS's powerful `defaults` system and user-friendly configuration management. Instead of remembering dozens of cryptic `defaults write` commands, you define your desired macOS settings in a clean TOML file and let fjrd handle the complex system interactions.
More features and additions to this tool are planned!

### Key Benefits

- **Declarative Configuration**: Define your desired state once, apply it anywhere
- **Version Control Friendly**: Track changes to your system configuration over time
- **Shareable**: Easily distribute configurations across teams or personal devices
- **Flexible Sources**: Load configs from local files, GitHub repos, or HTTPS URLs
- **Safe**: Built-in safety checks and user approval for potentially risky changes
- **Comprehensive**: Supports all major macOS system areas with room for custom settings

## How It Works

fjrd operates by translating your TOML configuration into macOS `defaults` commands, which modify the system preference files (`.plist` files) stored in `~/Library/Preferences/`. After applying settings, fjrd automatically restarts relevant system processes to ensure changes take effect immediately.

### The Process

1. **Parse Configuration**: Reads your TOML file from local path, GitHub repo, or URL
2. **Validate Settings**: Ensures all values are within acceptable ranges and formats
3. **Generate Commands**: Creates appropriate `defaults write` commands for each setting
4. **Apply Changes**: Executes commands to modify system preference files
5. **Restart Processes**: Kills and restarts affected system processes (Dock, Finder, etc.)

## Installation

### Install from GitHub

```bash
go install github.com/RATIU5/fjrd/cmd/fjrd@latest
```

### Build from Source

```bash
git clone https://github.com/RATIU5/fjrd
cd fjrd
go build ./cmd/fjrd
```

## Quick Start

### Basic Usage

```bash
# Apply local configuration file
fjrd config.toml

# Apply config from GitHub repository (uses fjrd.toml from repo root)
fjrd username/repository-name

# Apply config from specific GitHub file
fjrd https://github.com/username/repo/blob/main/my-config.toml

# Apply config from any HTTPS URL
fjrd https://example.com/macos-config.toml
```

### Simple Configuration Example

Create a `config.toml` file:

```toml
version = 1

[macos.dock]
autohide = true
orientation = "left"
tilesize = 48

[macos.finder]
show-all-extensions = true
show-path-bar = true
preferred-view-style = "column"

[macos.screenshots]
disable-shadow = true
format = "png"
save-location = "~/Desktop/Screenshots"
```

Then apply it:

```bash
fjrd config.toml
```

## Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-log-level` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `-log-format` | `text` | Log format: `text`, `json` |
| `-verbose` | `false` | Enable debug logging (same as `-log-level=debug`) |
| `-quiet` | `false` | Suppress non-error output |
| `-timeout` | `30s` | Operation timeout (e.g., `60s`, `2m`) |
| `-help` | `false` | Show help message |

### Examples

```bash
# Verbose logging for troubleshooting
fjrd -verbose config.toml

# JSON logs for automated processing
fjrd -log-format=json config.toml

# Quiet mode for scripts
fjrd -quiet config.toml

# Extended timeout for slow operations
fjrd -timeout=60s config.toml
```

## Configuration Reference

### Configuration Structure

All fjrd configuration files follow this basic structure:

```toml
version = 1

[macos.desktop]
# Desktop settings

[macos.dock]
# Dock settings

[macos.finder]
# Finder settings

# ... other system areas

[macos.defaultsRaw]
# Direct access to any defaults setting
```

### System Areas

#### Desktop Settings (`[macos.desktop]`)

Controls desktop appearance and behavior.

| Property | Type | Description | System Effect |
|----------|------|-------------|---------------|
| `sort-folders-first` | bool | Keep folders on top when sorting on desktop | Modifies `com.apple.finder._FXSortFoldersFirstOnDesktop` |
| `show-icons` | bool | Show/hide all desktop icons | Modifies `com.apple.finder.CreateDesktop` |
| `show-hard-drives` | bool | Show internal hard drives on desktop | Modifies `com.apple.finder.ShowHardDrivesOnDesktop` |
| `show-external-hard-drives` | bool | Show external hard drives on desktop | Modifies `com.apple.finder.ShowExternalHardDrivesOnDesktop` |
| `show-removable-media` | bool | Show removable media (USB drives) on desktop | Modifies `com.apple.finder.ShowRemovableMediaOnDesktop` |
| `show-mounted-servers` | bool | Show mounted network servers on desktop | Modifies `com.apple.finder.ShowMountedServersOnDesktop` |

#### Dock Settings (`[macos.dock]`)

Controls Dock appearance, behavior, and animations.

| Property | Type | Description | System Effect |
|----------|------|-------------|---------------|
| `autohide` | bool | Enable/disable Dock auto-hiding | Modifies `com.apple.dock.autohide` |
| `orientation` | string | Set Dock position: `left`, `bottom`, `right` | Modifies `com.apple.dock.orientation` |
| `tilesize` | int | Set icon size in pixels (e.g., `48`) | Modifies `com.apple.dock.tilesize` |
| `autohide-time` | float | Animation duration for showing/hiding (e.g., `0.5`) | Modifies `com.apple.dock.autohide-time-modifier` |
| `autohide-delay` | float | Delay before hiding starts (e.g., `0.2`) | Modifies `com.apple.dock.autohide-delay` |
| `show-recents` | bool | Show recent applications in Dock | Modifies `com.apple.dock.show-recents` |
| `min-effect` | string | Window minimize effect: `genie`, `scale`, `suck` | Modifies `com.apple.dock.mineffect` |
| `static-only` | bool | Only show running applications | Modifies `com.apple.dock.static-only` |
| `scroll-to-open` | bool | Scroll on Dock icon opens Exposé | Modifies `com.apple.dock.scroll-to-open` |

#### Finder Settings (`[macos.finder]`)

Controls Finder behavior and appearance.

| Property | Type | Description | System Effect |
|----------|------|-------------|---------------|
| `show-all-extensions` | bool | Show all file extensions | Modifies `NSGlobalDomain.AppleShowAllExtensions` |
| `show-all-files` | bool | Show hidden files (starting with .) | Modifies `com.apple.finder.AppleShowAllFiles` |
| `show-path-bar` | bool | Show path bar at bottom of windows | Modifies `com.apple.finder.ShowPathbar` |
| `preferred-view-style` | string | Default view: `column`, `list`, `gallery`, `icon` | Modifies `com.apple.finder.FXPreferredViewStyle` |
| `sort-folders-first` | bool | Keep folders on top when sorting | Modifies `com.apple.finder._FXSortFoldersFirst` |
| `finder-spawn-tab` | bool | Open new windows as tabs | Modifies `com.apple.finder.FinderSpawnTab` |
| `default-search-scope` | string | Search scope: `current`, `previous`, `mac` | Modifies `com.apple.finder.FXDefaultSearchScope` |
| `remove-old-trash-items` | bool | Auto-remove trash items after 30 days | Modifies `com.apple.finder.FXRemoveOldTrashItems` |
| `show-extension-change-warning` | bool | Warn when changing file extensions | Modifies `com.apple.finder.FXEnableExtensionChangeWarning` |
| `save-new-docs-to-cloud` | bool | Save new documents to iCloud by default | Modifies `NSGlobalDomain.NSDocumentSaveNewDocumentsToCloud` |

#### Keyboard Settings (`[macos.keyboard]`)

Controls keyboard behavior and shortcuts.

| Property | Type | Description | System Effect |
|----------|------|-------------|---------------|
| `key-hold-shows-accents` | bool | Press and hold shows accent marks | Modifies `NSGlobalDomain.ApplePressAndHoldEnabled` |
| `fn-key-behavior` | string | Fn key behavior: `dictation`, `input-source`, `emoji`, `none` | Modifies `com.apple.HIToolbox.AppleFnUsageType` |
| `special-f-keys` | bool | Use F1, F2, etc. as standard function keys | Modifies `NSGlobalDomain.com.apple.keyboard.fnState` |
| `tab-navigation` | bool | Tab navigates to all controls in dialogs | Modifies `NSGlobalDomain.AppleKeyboardUIMode` |

#### Menu Bar Settings (`[macos.menubar]`)

Controls menu bar appearance and behavior.

| Property | Type | Description | System Effect |
|----------|------|-------------|---------------|
| `clock-flash-date-separators` | bool | Flash date separators in menu bar clock | Modifies `com.apple.menuextra.clock.FlashDateSeparators` |
| `clock-date-format` | string | Date/time format (e.g., `"EEE d MMM HH:mm:ss"`) | Modifies `com.apple.menuextra.clock.DateFormat` |

#### Mission Control Settings (`[macos.missionControl]`)

Controls Spaces and Mission Control behavior.

| Property | Type | Description | System Effect |
|----------|------|-------------|---------------|
| `auto-rearrange-spaces` | bool | Auto-rearrange Spaces by recent use | Modifies `com.apple.dock.mru-spaces` |
| `group-windows-by-app` | bool | Group windows by app in Mission Control | Modifies `com.apple.dock.expose-group-apps` |
| `switch-to-apps-open-window` | bool | Switch to Space with app's open windows | Modifies `NSGlobalDomain.AppleSpacesSwitchOnActivate` |
| `displays-have-separate-spaces` | bool | Each display has separate Spaces | Modifies `com.apple.spaces.spans-displays` |

#### Mouse Settings (`[macos.mouse]`)

Controls mouse behavior and sensitivity.

| Property | Type | Description | System Effect |
|----------|------|-------------|---------------|
| `acceleration` | bool | Enable/disable mouse acceleration | Modifies `NSGlobalDomain.com.apple.mouse.linear` |
| `speed` | float | Mouse tracking speed (e.g., `1.5`) | Modifies `NSGlobalDomain.com.apple.mouse.scaling` |

#### Safari Settings (`[macos.safari]`)

Controls Safari browser behavior.

| Property | Type | Description | System Effect |
|----------|------|-------------|---------------|
| `show-full-url` | bool | Show full URL in address bar | Modifies `com.apple.Safari.ShowFullURLInSmartSearchField` |

#### Screenshot Settings (`[macos.screenshots]`)

Controls screenshot behavior and formatting.

| Property | Type | Description | System Effect |
|----------|------|-------------|---------------|
| `disable-shadow` | bool | Disable shadow effect for window screenshots | Modifies `com.apple.screencapture.disable-shadow` |
| `include-date` | bool | Include date in screenshot filename | Modifies `com.apple.screencapture.include-date` |
| `save-location` | string | Default save location (e.g., `~/Desktop`) | Modifies `com.apple.screencapture.location` |
| `show-thumbnail` | bool | Show floating thumbnail after capture | Modifies `com.apple.screencapture.show-thumbnail` |
| `format` | string | Image format: `png`, `jpg`, `pdf`, `gif`, etc. | Modifies `com.apple.screencapture.type` |

#### Trackpad Settings (`[macos.trackpad]`)

Controls trackpad behavior and gestures.

| Property | Type | Description | System Effect |
|----------|------|-------------|---------------|
| `click-weight` | int | Click pressure sensitivity: `0`-`3` | Modifies `com.apple.AppleMultitouchTrackpad.FirstClickThreshold` |
| `three-finger-drag` | bool | Enable three-finger drag gesture | Modifies `com.apple.AppleMultitouchTrackpad.TrackpadThreeFingerDrag` |

## Advanced Configuration

### Raw Defaults Access (`[macos.defaultsRaw]`)

The `defaultsRaw` section provides direct access to any macOS defaults setting, even those not explicitly supported by fjrd's schema. This is powerful but requires careful use.

```toml
[macos.defaultsRaw]
"com.apple.dock.workspaces-auto-swoosh" = { value = 0, type = "int" }
"com.apple.finder.NewWindowTarget" = { value = "PfDe", type = "string" }
"com.apple.screencapture.show-thumbnail" = { value = false, type = "bool" }
```

#### Value Types

| Type | Description | Example |
|------|-------------|---------|
| `"string"` | Text values | `{ value = "PfDe", type = "string" }` |
| `"bool"` | Boolean values | `{ value = true, type = "bool" }` |
| `"int"` | Integer numbers | `{ value = 42, type = "int" }` |
| `"float"` | Decimal numbers | `{ value = 1.5, type = "float" }` |

#### Safety Features

- **User Approval**: fjrd will list all raw defaults and ask for confirmation before applying
- **Validation**: Values are validated against their specified types
- **Reversible**: Settings can be reset to system defaults

#### Resetting to Defaults

```toml
[macos.defaultsRaw]
# Reset to system default
"com.apple.dock.workspaces-auto-swoosh" = { value = "default", type = "int" }

# Explicit reset flag
"com.apple.dock.expose-animation-duration" = { value = 0.1, type = "float", reset = true }
```

When reset, fjrd runs `defaults delete domain.key` to restore the system default.

## Complete Configuration Example

```toml
version = 1

[macos.desktop]
sort-folders-first = true
show-icons = true
show-hard-drives = false
show-external-hard-drives = true
show-removable-media = true
show-mounted-servers = false

[macos.dock]
autohide = true
orientation = "left"
tilesize = 48
autohide-time = 0.3
autohide-delay = 0.0
show-recents = false
min-effect = "scale"
static-only = false
scroll-to-open = true

[macos.finder]
show-all-extensions = true
show-all-files = false
show-path-bar = true
preferred-view-style = "column"
sort-folders-first = true
finder-spawn-tab = true
default-search-scope = "current"
remove-old-trash-items = true
show-extension-change-warning = true
save-new-docs-to-cloud = false

[macos.keyboard]
key-hold-shows-accents = true
fn-key-behavior = "dictation"
special-f-keys = false
tab-navigation = true

[macos.menubar]
clock-flash-date-separators = false
clock-date-format = "EEE d MMM HH:mm:ss"

[macos.missionControl]
auto-rearrange-spaces = false
group-windows-by-app = true
switch-to-apps-open-window = true
displays-have-separate-spaces = true

[macos.mouse]
acceleration = false
speed = 2.0

[macos.safari]
show-full-url = true

[macos.screenshots]
disable-shadow = true
include-date = true
save-location = "~/Desktop/Screenshots"
show-thumbnail = true
format = "png"

[macos.trackpad]
click-weight = 1
three-finger-drag = true
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

[See the LICENSE](LICENSE.md)
