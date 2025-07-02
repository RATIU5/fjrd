package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/macos/desktop"
	"github.com/RATIU5/fjrd/internal/macos/dock"
	"github.com/RATIU5/fjrd/internal/macos/finder"
	"github.com/RATIU5/fjrd/internal/macos/safari"
	"github.com/RATIU5/fjrd/internal/macos/screenshots"
)

type MacosConfig struct {
	Dock        dock.Config        `toml:"dock"`
	Finder      finder.Config      `toml:"finder"`
	Desktop     desktop.Config     `toml:"desktop"`
	Safari      safari.Config      `toml:"safari"`
	Screenshots screenshots.Config `toml:"screenshots"`
	DefaultsRaw defaults.Raw       `toml:"defaultsRaw"`
}

type FjrdConfig struct {
	Version Version     `toml:"version"`
	Macos   MacosConfig `toml:"macos"`
}

func (m *MacosConfig) String() string {
	defaultsStr := m.DefaultsRaw.String()
	defaultsLines := strings.Split(defaultsStr, "\n")
	for i, line := range defaultsLines {
		if i > 0 && i < len(defaultsLines)-1 {
			defaultsLines[i] = "  " + line
		}
	}
	indentedDefaults := strings.Join(defaultsLines, "\n")

	return fmt.Sprintf("Macos{\n  dock: %s\n  defaultsRaw: %s\n}",
		m.Dock.String(),
		indentedDefaults)
}

func (c *FjrdConfig) String() string {
	macosStr := c.Macos.String()
	macosLines := strings.Split(macosStr, "\n")
	for i, line := range macosLines {
		if i > 0 && i < len(macosLines)-1 {
			macosLines[i] = "  " + line
		}
	}
	indentedMacos := strings.Join(macosLines, "\n")

	return fmt.Sprintf("FjrdConfig{\n  version: %d\n  macos: %s\n}",
		c.Version,
		indentedMacos)
}

func (c *MacosConfig) Validate() error {
	if err := c.Dock.Validate(); err != nil {
		return err
	}
	if err := c.Finder.Validate(); err != nil {
		return err
	}
	if err := c.Desktop.Validate(); err != nil {
		return err
	}
	if err := c.Safari.Validate(); err != nil {
		return err
	}
	if err := c.Screenshots.Validate(); err != nil {
		return err
	}
	if err := c.DefaultsRaw.Validate(); err != nil {
		return err
	}
	return nil
}

func (c *FjrdConfig) Validate() error {
	if err := c.Version.Validate(); err != nil {
		return err
	}
	if err := c.Macos.Validate(); err != nil {
		return err
	}
	return nil
}

func (c *MacosConfig) Execute(ctx context.Context, log interface {
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}) error {
	log.Info("Executing macOS configuration")

	log.Info("Applying dock configuration")
	if err := c.Dock.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute dock configuration: %w", err)
	}

	log.Info("Applying finder configuration")
	if err := c.Finder.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute finder configuration: %w", err)
	}

	log.Info("Applying desktop configuration")
	if err := c.Finder.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute desktop configuration: %w", err)
	}

	log.Info("Applying safari configuration")
	if err := c.Safari.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute safari configuration: %w", err)
	}

	log.Info("Applying screenshots configuration")
	if err := c.Screenshots.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute screenshots configuration: %w", err)
	}

	log.Info("Applying raw defaults")
	if err := c.DefaultsRaw.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute raw defaults: %w", err)
	}

	log.Info("macOS configuration applied successfully")
	return nil
}

func (c *FjrdConfig) Execute(ctx context.Context, log interface {
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}) error {
	log.Info("Executing fjrd configuration", "version", c.Version)
	if err := c.Macos.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute macOS configuration: %w", err)
	}
	return nil
}
