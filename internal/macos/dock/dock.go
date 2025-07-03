package dock

import (
	"context"
	"fmt"

	"github.com/RATIU5/fjrd/internal/errors"
	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/shared"
)

type Config struct {
	Autohide      *bool      `toml:"autohide,omitempty"`
	Orientation   *Position  `toml:"orientation,omitempty"`
	TileSize      *int16     `toml:"tilesize,omitempty"`
	AutohideTime  *float32   `toml:"autohide-time,omitempty"`
	AutohideDelay *float32   `toml:"autohide-delay,omitempty"`
	ShowRecents   *bool      `toml:"show-recents,omitempty"`
	MinEffect     *MinEffect `toml:"min-effect,omitempty"`
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
	return shared.FormatConfig("Dock", d)
}

func (d *Config) Fields() map[string]any {
	fields := make(map[string]any)

	if d.Autohide != nil {
		fields["autohide"] = d.Autohide
	}
	if d.Orientation != nil {
		fields["orientation"] = d.Orientation
	}
	if d.TileSize != nil {
		fields["tilesize"] = d.TileSize
	}
	if d.AutohideTime != nil {
		fields["autohide-time"] = d.AutohideTime
	}
	if d.AutohideDelay != nil {
		fields["autohide-delay"] = d.AutohideDelay
	}
	if d.ShowRecents != nil {
		fields["show-recents"] = d.ShowRecents
	}
	if d.MinEffect != nil {
		fields["min-effect"] = d.MinEffect
	}
	if d.StaticOnly != nil {
		fields["static-only"] = d.StaticOnly
	}
	if d.ScrollToOpen != nil {
		fields["scroll-to-open"] = d.ScrollToOpen
	}

	return fields
}

func (d *Config) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("dock")
	log.Debug("Configuring dock settings")

	batch := defaults.NewBatchExecutor()
	const dockDomain = "com.apple.dock"

	multiErr := errors.NewMultiError()

	if d.Autohide != nil {
		batch.AddBool(dockDomain, "autohide", *d.Autohide)
	}

	if d.Orientation != nil {
		orientationValue := defaults.NewEnumValue(d.Orientation.String(), []string{"left", "bottom", "right"})
		batch.AddCommand(defaults.Command{
			Domain: dockDomain,
			Key:    "orientation",
			Value:  orientationValue,
		})
	}

	if d.TileSize != nil {
		if err := batch.AddInt(dockDomain, "tilesize", *d.TileSize); err != nil {
			multiErr.Add(errors.WrapConfigError("dock", "add_command", "tilesize", *d.TileSize, err))
		}
	}

	if d.AutohideTime != nil {
		if err := batch.AddFloat(dockDomain, "autohide-time-modifier", *d.AutohideTime); err != nil {
			multiErr.Add(errors.WrapConfigError("dock", "add_command", "autohide-time-modifier", *d.AutohideTime, err))
		}
	}

	if d.AutohideDelay != nil {
		if err := batch.AddFloat(dockDomain, "autohide-delay", *d.AutohideDelay); err != nil {
			multiErr.Add(errors.WrapConfigError("dock", "add_command", "autohide-delay", *d.AutohideDelay, err))
		}
	}

	if d.ShowRecents != nil {
		batch.AddBool(dockDomain, "show-recents", *d.ShowRecents)
	}

	if d.MinEffect != nil {
		minEffectValue := defaults.NewEnumValue(d.MinEffect.String(), []string{"genie", "scale", "suck"})
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

	if err := multiErr.ToError(); err != nil {
		return err
	}

	log.Debug("Applying dock defaults")
	if err := batch.Execute(ctx, log); err != nil {
		return errors.WrapConfigError("dock", "execute_batch", "", nil, err)
	}

	log.Debug("Restarting dock to apply changes")
	killall := defaults.NewKillallExecutor("Dock")
	if err := killall.Execute(ctx); err != nil {
		return errors.WrapConfigError("dock", "restart_process", "Dock", nil, err)
	}

	log.Debug("Dock configuration applied successfully")
	return nil
}
