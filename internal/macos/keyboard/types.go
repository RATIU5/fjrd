package keyboard

import (
	"fmt"
	"strings"
)

type FnBehavior int

const (
	None FnBehavior = iota
	InputSource
	Emoji
	Dictation
)

func (b FnBehavior) IsValid() bool {
	switch b {
	case None, InputSource, Emoji, Dictation:
		return true
	default:
		return false
	}
}

func (b FnBehavior) String() string {
	switch b {
	case Dictation:
		return "dictation"
	case InputSource:
		return "input-source"
	case Emoji:
		return "emoji"
	default:
		return "none"
	}
}

func (b *FnBehavior) UnmarshalText(text []byte) error {
	parsed, err := ParseFnBehavior(string(text))
	if err != nil {
		return err
	}
	*b = parsed
	return nil
}

func (b FnBehavior) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func ParseFnBehavior(s string) (FnBehavior, error) {
	switch strings.ToLower(s) {
	case "dictation":
		return Dictation, nil
	case "input-source":
		return InputSource, nil
	case "emoji":
		return Emoji, nil
	case "none":
		return None, nil
	default:
		return None, fmt.Errorf("invalid fn-key-behavior %q, must be one of dictation, input source, emoji, none", s)
	}
}

func AllFnBehaviors() []FnBehavior {
	return []FnBehavior{Dictation, InputSource, Emoji, None}
}

type ValueConverter interface {
	Convert() any
}

type TabNavigationValue struct {
	Enabled bool
}

func NewTabNavigationValue(enabled bool) *TabNavigationValue {
	return &TabNavigationValue{Enabled: enabled}
}

func (t *TabNavigationValue) Convert() any {
	if t.Enabled {
		return 2
	}
	return 0
}
