package dock

import (
	"context"
	"fmt"
	"strings"

	"github.com/RATIU5/fjrd/internal/macos/defaults"
)

type Config struct {
	Autohide      *bool      `toml:"autohide,omitempty"`
	Orientation   *Position  `toml:"orientation,omitempty"`
	TileSize      *int16     `toml:"tilesize,omitempty"`
	AutohideTime  *float32   `toml:"autohide-time-modifier,omitempty"`
	AutohideDelay *float32   `toml:"autohide-delay,omitempty"`
	ShowRecents   *bool      `toml:"show-recents,omitempty"`
	MinEffect     *MinEffect `toml:"mineffect,omitempty"`
	StaticOnly    *bool      `toml:"static-only,omitempty"`
	ScrollToOpen  *bool      `toml:"scroll-to-open,omitempty"`
}

func (d *Config) Validate() error {
	if d.Orientation != nil && !d.Orientation.IsValid() {
		return fmt.Errorf("invalid orientation: %s", *d.Orientation)
	}
	if d.MinEffect != nil && !d.MinEffect.IsValid() {
		return fmt.Errorf("invalid minimize effect: %s", *d.MinEffect)
	}
	return nil
}

func (d *Config) String() string {
	var parts []string

	if d.Autohide != nil {
		parts = append(parts, fmt.Sprintf("autohide: %t", *d.Autohide))
	}
	if d.Orientation != nil {
		parts = append(parts, fmt.Sprintf("orientation: %s", *d.Orientation))
	}
	if d.TileSize != nil {
		parts = append(parts, fmt.Sprintf("tilesize: %d", *d.TileSize))
	}
	if d.AutohideTime != nil {
		parts = append(parts, fmt.Sprintf("autohide-time: %.2f", *d.AutohideTime))
	}
	if d.AutohideDelay != nil {
		parts = append(parts, fmt.Sprintf("autohide-delay: %.2f", *d.AutohideDelay))
	}
	if d.ShowRecents != nil {
		parts = append(parts, fmt.Sprintf("show-recents: %t", *d.ShowRecents))
	}
	if d.MinEffect != nil {
		parts = append(parts, fmt.Sprintf("mineffect: %s", *d.MinEffect))
	}
	if d.StaticOnly != nil {
		parts = append(parts, fmt.Sprintf("static-only: %t", *d.StaticOnly))
	}
	if d.ScrollToOpen != nil {
		parts = append(parts, fmt.Sprintf("scroll-to-open: %t", *d.ScrollToOpen))
	}

	if len(parts) == 0 {
		return "Dock{}"
	}

	return fmt.Sprintf("Dock{%s}", strings.Join(parts, ", "))
}

func (d *Config) Execute(ctx context.Context, log interface {
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}) error {
	log.Debug("Configuring dock settings")
	batch := defaults.NewBatchExecutor()
	const dockDomain = "com.apple.dock"

	if d.Autohide != nil {
		batch.AddBool(dockDomain, "autohide", *d.Autohide)
	}

	if d.Orientation != nil {
		orientationValue := defaults.NewEnumValue(string(*d.Orientation), []string{"left", "bottom", "right"})
		batch.AddCommand(defaults.Command{
			Domain: dockDomain,
			Key:    "orientation",
			Value:  orientationValue,
		})
	}

	if d.TileSize != nil {
		if err := batch.AddInt(dockDomain, "tilesize", *d.TileSize); err != nil {
			return fmt.Errorf("failed to add tilesize command: %w", err)
		}
	}

	if d.AutohideTime != nil {
		if err := batch.AddFloat(dockDomain, "autohide-time-modifier", *d.AutohideTime); err != nil {
			return fmt.Errorf("failed to add autohide-time-modifier command: %w", err)
		}
	}

	if d.AutohideDelay != nil {
		if err := batch.AddFloat(dockDomain, "autohide-delay", *d.AutohideDelay); err != nil {
			return fmt.Errorf("failed to add autohide-delay command: %w", err)
		}
	}

	if d.ShowRecents != nil {
		batch.AddBool(dockDomain, "show-recents", *d.ShowRecents)
	}

	if d.MinEffect != nil {
		minEffectValue := defaults.NewEnumValue(string(*d.MinEffect), []string{"genie", "scale", "suck"})
		batch.AddCommand(defaults.Command{
			Domain: dockDomain,
			Key:    "mineffect",
			Value:  minEffectValue,
		})
	}

	if d.StaticOnly != nil {
		batch.AddBool(dockDomain, "static-only", *d.StaticOnly)
	}

	if d.ScrollToOpen != nil {
		batch.AddBool(dockDomain, "scroll-to-open", *d.ScrollToOpen)
	}

	log.Debug("Applying dock defaults")
	if err := batch.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute dock configuration: %w", err)
	}

	log.Debug("Restarting dock to apply changes")
	killall := defaults.NewKillallExecutor("Dock")
	if err := killall.Execute(ctx); err != nil {
		return fmt.Errorf("failed to restart dock: %w", err)
	}

	log.Info("Dock configuration applied successfully")
	return nil
}
