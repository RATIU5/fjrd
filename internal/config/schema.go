package config

import (
	"context"
	"fmt"

	"github.com/RATIU5/fjrd/internal/errors"
	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/macos/desktop"
	"github.com/RATIU5/fjrd/internal/macos/dock"
	"github.com/RATIU5/fjrd/internal/macos/finder"
	"github.com/RATIU5/fjrd/internal/macos/keyboard"
	"github.com/RATIU5/fjrd/internal/macos/menubar"
	"github.com/RATIU5/fjrd/internal/macos/missionControl"
	"github.com/RATIU5/fjrd/internal/macos/mouse"
	"github.com/RATIU5/fjrd/internal/macos/safari"
	"github.com/RATIU5/fjrd/internal/macos/screenshots"
	"github.com/RATIU5/fjrd/internal/macos/trackpad"
	"github.com/RATIU5/fjrd/internal/shared"
)

type MacosConfig struct {
	Dock           dock.Config           `toml:"dock"`
	Finder         finder.Config         `toml:"finder"`
	Desktop        desktop.Config        `toml:"desktop"`
	Safari         safari.Config         `toml:"safari"`
	Screenshots    screenshots.Config    `toml:"screenshots"`
	Menubar        menubar.Config        `toml:"meubar"`
	Mouse          mouse.Config          `toml:"mouse"`
	Trackpad       trackpad.Config       `toml:"trackpad"`
	Keyboard       keyboard.Config       `toml:"keyboard"`
	MissionControl missionControl.Config `toml:"mission-control"`
	DefaultsRaw    defaults.Raw          `toml:"defaultsRaw"`
}

type FjrdConfig struct {
	Version Version     `toml:"version"`
	Macos   MacosConfig `toml:"macos"`
}

func (m *MacosConfig) String() string {
	return shared.FormatConfig("Macos", m)
}

func (m *MacosConfig) Fields() map[string]any {
	return map[string]any{
		"dock":            m.Dock,
		"finder":          m.Finder,
		"desktop":         m.Desktop,
		"safari":          m.Safari,
		"screenshots":     m.Screenshots,
		"menubar":         m.Menubar,
		"mouse":           m.Mouse,
		"trackpad":        m.Trackpad,
		"keyboard":        m.Keyboard,
		"mission-control": m.MissionControl,
		"defaultsRaw":     m.DefaultsRaw,
	}
}

func (c *FjrdConfig) String() string {
	return shared.FormatConfig("FjrdConfig", c)
}

func (c *FjrdConfig) Fields() map[string]any {
	return map[string]any{
		"version": c.Version,
		"macos":   c.Macos,
	}
}

func (c *MacosConfig) Validate() error {
	return shared.ValidateAll(
		&c.Dock,
		&c.Finder,
		&c.Desktop,
		&c.Safari,
		&c.Screenshots,
		&c.Menubar,
		&c.Mouse,
		&c.Trackpad,
		&c.Keyboard,
		&c.MissionControl,
		&c.DefaultsRaw,
	)
}

func (c *FjrdConfig) Validate() error {
	return shared.ValidateAll(&c.Version, &c.Macos)
}

func (c *MacosConfig) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("macos")
	multiErr := errors.NewMultiError()

	if err := c.Dock.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("macos", "execute", "dock", nil, err))
	}

	if err := c.Finder.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("macos", "execute", "finder", nil, err))
	}

	if err := c.Desktop.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("macos", "execute", "desktop", nil, err))
	}

	if err := c.Safari.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("macos", "execute", "safari", nil, err))
	}

	if err := c.Screenshots.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("macos", "execute", "screenshots", nil, err))
	}

	if err := c.Menubar.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("macos", "execute", "menubar", nil, err))
	}

	if err := c.Mouse.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("macos", "execute", "mouse", nil, err))
	}

	if err := c.Trackpad.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("macos", "execute", "trackpad", nil, err))
	}

	if err := c.Keyboard.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("macos", "execute", "keyboard", nil, err))
	}

	if err := c.MissionControl.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("macos", "execute", "mission-control", nil, err))
	}

	if err := c.DefaultsRaw.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("macos", "execute", "defaultsRaw", nil, err))
	}

	if err := multiErr.ToError(); err != nil {
		return err
	}

	log.Debug("macos configuration applied successfully")
	return nil
}

func (c *FjrdConfig) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("fjrd")
	log.Debug("Executing fjrd configuration", "version", c.Version)
	if err := c.Macos.Execute(ctx, log); err != nil {
		return errors.WrapConfigError("fjrd", "execute", "macos", nil, err)
	}
	return nil
}

func (c *FjrdConfig) RequiresRawDefaultsApproval() bool {
	return len(c.Macos.DefaultsRaw) > 0
}

func (c *FjrdConfig) ListRawDefaults() []string {
	var commands []string
	for domainKey, entry := range c.Macos.DefaultsRaw {
		var cmdStr string
		if entry.ShouldReset() {
			cmdStr = fmt.Sprintf("defaults delete %s", domainKey)
		} else {
			switch entry.Type {
			case defaults.TypeString:
				if v, ok := entry.GetStringValue(); ok {
					cmdStr = fmt.Sprintf("defaults write %s -string \"%s\"", domainKey, v)
				}
			case defaults.TypeBool:
				if v, ok := entry.GetBoolValue(); ok {
					cmdStr = fmt.Sprintf("defaults write %s -bool %t", domainKey, v)
				}
			case defaults.TypeInt:
				if v, ok := entry.GetIntValue(); ok {
					cmdStr = fmt.Sprintf("defaults write %s -int %d", domainKey, v)
				}
			case defaults.TypeFloat:
				if v, ok := entry.GetFloatValue(); ok {
					cmdStr = fmt.Sprintf("defaults write %s -float %f", domainKey, v)
				}
			}
		}
		commands = append(commands, cmdStr)
	}
	return commands
}
