package dock

import (
	"fmt"
	"strings"
)

type Position string
type MinEffect string

const (
	PositionLeft   Position = "left"
	PositionBottom Position = "bottom"
	PositionRight  Position = "right"
)

const (
	EffectGenie MinEffect = "genie"
	EffectScale MinEffect = "scale"
	EffectSuck  MinEffect = "suck"
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
	return string(p)
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
	return string(e)
}

func ParsePosition(s string) (Position, error) {
	pos := Position(strings.ToLower(s))
	if !pos.IsValid() {
		return "", fmt.Errorf("invalid dock position %q, must be one of: left, bottom, right", s)
	}
	return pos, nil
}

func ParseMinEffect(s string) (MinEffect, error) {
	effect := MinEffect(strings.ToLower(s))
	if !effect.IsValid() {
		return "", fmt.Errorf("invalid dock effect %q, must be one of: genie, scale, suck", s)
	}
	return effect, nil
}

func AllPositions() []Position {
	return []Position{PositionLeft, PositionBottom, PositionRight}
}

func AllMinEffects() []MinEffect {
	return []MinEffect{EffectGenie, EffectScale, EffectSuck}
}