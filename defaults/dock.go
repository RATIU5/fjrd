package defaults

import (
	"fmt"
	"strings"
)

type DockPosition string
type DockMinEffect string

const (
	PositionLeft   DockPosition = "left"
	PositionBottom DockPosition = "bottom"
	PositionRight  DockPosition = "right"
)

const (
	EffectGenie DockMinEffect = "genie"
	EffectScale DockMinEffect = "scale"
	EffectSuck  DockMinEffect = "suck"
)

func (dp DockPosition) IsValid() bool {
	switch dp {
	case PositionLeft, PositionBottom, PositionRight:
		return true
	default:
		return false
	}
}

func (dp DockPosition) String() string {
	return string(dp)
}

func (de DockMinEffect) IsValid() bool {
	switch de {
	case EffectGenie, EffectScale, EffectSuck:
		return true
	default:
		return false
	}
}

func (de DockMinEffect) String() string {
	return string(de)
}

func ParseDockPosition(s string) (DockPosition, error) {
	pos := DockPosition(strings.ToLower(s))
	if !pos.IsValid() {
		return "", fmt.Errorf("invalid dock position %q, must be one of: left, bottom, right", s)
	}
	return pos, nil
}

func ParseDockMinEffect(s string) (DockMinEffect, error) {
	effect := DockMinEffect(strings.ToLower(s))
	if !effect.IsValid() {
		return "", fmt.Errorf("invalid dock effect %q, must be one of: genie, scale, suck", s)
	}
	return effect, nil
}

func AllPositions() []DockPosition {
	return []DockPosition{PositionLeft, PositionBottom, PositionRight}
}

func AllMinEffects() []DockMinEffect {
	return []DockMinEffect{EffectGenie, EffectScale, EffectSuck}
}

type Dock struct {
	Autohide      *bool          `toml:"autohide,omitempty" comment:"should dock autohide"`
	Orientation   *DockPosition  `toml:"orientation,omitempty" comment:"dock position"`
	TileSize      *int16         `toml:"tilesize,omitempty" comment:"icon size"`
	AutohideTime  *float32       `toml:"autohide-time-modifier,omitempty" comment:"autohide time"`
	AutohideDelay *float32       `toml:"autohide-delay,omitempty" comment:"autohide delay"`
	ShowRecents   *bool          `toml:"show-recents,omitempty" comment:"show recent apps"`
	MinEffect     *DockMinEffect `toml:"mineffect,omitempty" comment:"minimize effect"`
	StaticOnly    *bool          `toml:"static-only,omitempty" comment:"active apps only"`
	ScrollToOpen  *bool          `toml:"scroll-to-open,omitempty" comment:"scroll to open app"`
}

func (d *Dock) Validate() error {
	if d.Orientation != nil && !d.Orientation.IsValid() {
		return fmt.Errorf("invalid orientation: %s", *d.Orientation)
	}
	if d.MinEffect != nil && !d.MinEffect.IsValid() {
		return fmt.Errorf("invalid minimize effect: %s", *d.MinEffect)
	}
	return nil
}

func (d *Dock) String() string {
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

func (d *Dock) Execute() error {
	batch := NewBatchExecutor()
	const dockDomain = "com.apple.dock"

	if d.Autohide != nil {
		batch.AddBool(dockDomain, "autohide", *d.Autohide)
	}

	if d.Orientation != nil {
		orientationValue := NewEnumValue(string(*d.Orientation), []string{"left", "bottom", "right"})
		batch.AddCommand(Command{
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
		minEffectValue := NewEnumValue(string(*d.MinEffect), []string{"genie", "scale", "suck"})
		batch.AddCommand(Command{
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

	if err := batch.Execute(); err != nil {
		return fmt.Errorf("failed to execute dock configuration: %w", err)
	}

	killall := NewKillallExecutor("Dock")
	if err := killall.Execute(); err != nil {
		return fmt.Errorf("failed to restart dock: %w", err)
	}

	return nil
}
