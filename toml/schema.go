package toml

import (
	"fmt"
	"strings"

	"github.com/RATIU5/fjrd/defaults"
)

type Macos struct {
	Dock        defaults.Dock `toml:"dock" comment:"dock settings"`
	DefaultsRaw defaults.Raw  `toml:"defaultsRaw" comment:"raw defaults queries"`
}

type FjrdConfig struct {
	Version Version `toml:"version" comment:"version of fjrd schema"`
	Macos   Macos   `toml:"macos" comment:"macos settings"`
}

func (m *Macos) String() string {
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

func (c *Macos) Validate() error {
	if err := c.Dock.Validate(); err != nil {
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

func (c *Macos) Execute() error {
	if err := c.Dock.Execute(); err != nil {
		return fmt.Errorf("failed to execute dock configuration: %w", err)
	}
	if err := c.DefaultsRaw.Execute(); err != nil {
		return fmt.Errorf("failed to execute raw defaults: %w", err)
	}
	return nil
}

func (c *FjrdConfig) Execute() error {
	if err := c.Macos.Execute(); err != nil {
		return fmt.Errorf("failed to execute macOS configuration: %w", err)
	}
	return nil
}
