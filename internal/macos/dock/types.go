package dock

import (
	"fmt"
	"strings"
)

type Position int
type MinEffect int

const (
	PositionLeft Position = iota
	PositionBottom
	PositionRight
)

const (
	EffectGenie MinEffect = iota
	EffectScale
	EffectSuck
)

func (p Position) IsValid() bool {
	switch p {
	case PositionLeft, PositionBottom, PositionRight:
		return true
	default:
		return false
	}
}

func (p Position) String() string {
	switch p {
	case PositionLeft:
		return "left"
	case PositionBottom:
		return "bottom"
	case PositionRight:
		return "right"
	default:
		return "bottom"
	}
}

func (p *Position) UnmarshalText(text []byte) error {
	parsed, err := ParsePosition(string(text))
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}

func (p Position) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (e MinEffect) IsValid() bool {
	switch e {
	case EffectGenie, EffectScale, EffectSuck:
		return true
	default:
		return false
	}
}

func (e MinEffect) String() string {
	switch e {
	case EffectGenie:
		return "genie"
	case EffectScale:
		return "scale"
	case EffectSuck:
		return "suck"
	default:
		return "genie"
	}
}

func (e *MinEffect) UnmarshalText(text []byte) error {
	parsed, err := ParseMinEffect(string(text))
	if err != nil {
		return err
	}
	*e = parsed
	return nil
}

func (e MinEffect) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

func ParsePosition(s string) (Position, error) {
	switch strings.ToLower(s) {
	case "left":
		return PositionLeft, nil
	case "bottom":
		return PositionBottom, nil
	case "right":
		return PositionRight, nil
	default:
		return PositionBottom, fmt.Errorf("invalid dock position %q, must be one of: left, bottom, right", s)
	}
}

func ParseMinEffect(s string) (MinEffect, error) {
	switch strings.ToLower(s) {
	case "genie":
		return EffectGenie, nil
	case "scale":
		return EffectScale, nil
	case "suck":
		return EffectSuck, nil
	default:
		return EffectGenie, fmt.Errorf("invalid dock effect %q, must be one of: genie, scale, suck", s)
	}
}

func AllPositions() []Position {
	return []Position{PositionLeft, PositionBottom, PositionRight}
}

func AllMinEffects() []MinEffect {
	return []MinEffect{EffectGenie, EffectScale, EffectSuck}
}