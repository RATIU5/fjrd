package defaults

import (
	"context"
	"fmt"
	"strings"
)

type DefaultsType string

const (
	TypeString DefaultsType = "string"
	TypeBool   DefaultsType = "bool"
	TypeInt    DefaultsType = "int"
	TypeFloat  DefaultsType = "float"
)

type RawEntry struct {
	RawValue any          `toml:"value"`
	Type     DefaultsType `toml:"type"`
	Reset    *bool        `toml:"reset,omitempty"`
}

func (e *RawEntry) GetStringValue() (string, bool) {
	if e.Type == TypeString {
		if v, ok := e.RawValue.(string); ok {
			return v, true
		}
	}
	return "", false
}

func (e *RawEntry) GetBoolValue() (bool, bool) {
	if e.Type == TypeBool {
		if e.IsDefaultString() {
			return false, true
		}
		if v, ok := e.RawValue.(bool); ok {
			return v, true
		}
	}
	return false, false
}

func (e *RawEntry) GetIntValue() (int16, bool) {
	if e.Type == TypeInt {
		if e.IsDefaultString() {
			return 0, true
		}
		switch v := e.RawValue.(type) {
		case int16:
			return v, true
		case int:
			return int16(v), true
		case int64:
			return int16(v), true
		}
	}
	return 0, false
}

func (e *RawEntry) GetFloatValue() (float32, bool) {
	if e.Type == TypeFloat {
		if e.IsDefaultString() {
			return 0.0, true
		}
		switch v := e.RawValue.(type) {
		case float32:
			return v, true
		case float64:
			return float32(v), true
		}
	}
	return 0., false
}

func (e *RawEntry) String() string {
	switch e.Type {
	case TypeString:
		if v, ok := e.GetStringValue(); ok {
			return fmt.Sprintf("string: %q", v)
		}
	case TypeBool:
		if v, ok := e.GetBoolValue(); ok {
			return fmt.Sprintf("bool: %t", v)
		}
	case TypeInt:
		if v, ok := e.GetIntValue(); ok {
			return fmt.Sprintf("int: %d", v)
		}
	case TypeFloat:
		if v, ok := e.GetFloatValue(); ok {
			return fmt.Sprintf("float: %f", v)
		}
	}
	return fmt.Sprintf("%s: <invalid>", e.Type)
}

func (e *RawEntry) IsReset() bool {
	return e.Reset != nil && *e.Reset
}

func (e *RawEntry) IsNullValue() bool {
	return e.RawValue == nil
}

func (e *RawEntry) IsDefaultString() bool {
	if v, ok := e.RawValue.(string); ok {
		return v == "default"
	}
	return false
}

func (e *RawEntry) ShouldReset() bool {
	return e.IsReset() || e.IsNullValue() || e.IsDefaultString()
}

func (e *RawEntry) Validate() error {
	if e.ShouldReset() {
		return nil
	}

	switch e.Type {
	case TypeString:
		if _, isValid := e.GetStringValue(); !isValid {
			return fmt.Errorf("%v is not of expected type string", e.RawValue)
		}
	case TypeInt:
		if _, isValid := e.GetIntValue(); !isValid {
			return fmt.Errorf("%v is not of expected type int", e.RawValue)
		}
	case TypeFloat:
		if _, isValid := e.GetFloatValue(); !isValid {
			return fmt.Errorf("%v is not of expected type float", e.RawValue)
		}
	case TypeBool:
		if _, isValid := e.GetBoolValue(); !isValid {
			return fmt.Errorf("%v is not of expected type bool", e.RawValue)
		}
	}
	return nil
}

type Raw map[string]RawEntry

func (r Raw) String() string {
	if len(r) == 0 {
		return "DefaultsRaw{}"
	}

	var parts []string
	for key, entry := range r {
		parts = append(parts, fmt.Sprintf("  %q: %s", key, entry.String()))
	}
	return fmt.Sprintf("DefaultsRaw{\n%s\n}", strings.Join(parts, "\n"))
}

func (r Raw) Validate() error {
	for _, value := range r {
		if err := value.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (r Raw) Execute(ctx context.Context, log interface {
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}) error {
	if len(r) == 0 {
		log.Debug("No raw defaults to apply")
		return nil
	}

	log.Debug("Processing raw defaults", "count", len(r))
	batch := NewBatchExecutor()

	for domainKey, entry := range r {
		log.Debug("Processing raw default", "domain_key", domainKey, "type", entry.Type)
		domainParts := strings.Split(domainKey, ".")
		if len(domainParts) < 2 {
			return fmt.Errorf("invalid domain format %s, expected format: com.apple.domain.key", domainKey)
		}

		key := domainParts[len(domainParts)-1]
		macosDomain := strings.Join(domainParts[:len(domainParts)-1], ".")
		log.Debug("Parsed domain and key", "domain", macosDomain, "key", key)

		var value Value
		var err error

		if entry.ShouldReset() {
			log.Debug("Creating reset value", "domain_key", domainKey, "type", entry.Type)
			switch entry.Type {
			case TypeString:
				value = NewResetStringValue()
			case TypeBool:
				value = NewResetBoolValue()
			case TypeInt:
				value = NewResetIntValue()
			case TypeFloat:
				value = NewResetFloatValue()
			default:
				return fmt.Errorf("unsupported type %s for reset of domain %s", entry.Type, domainKey)
			}
		} else {
			switch entry.Type {
			case TypeString:
				if v, ok := entry.GetStringValue(); ok {
					value = NewStringValue(v)
				} else {
					return fmt.Errorf("invalid string value for domain %s", domainKey)
				}
			case TypeBool:
				if v, ok := entry.GetBoolValue(); ok {
					value = NewBoolValue(v)
				} else {
					return fmt.Errorf("invalid bool value for domain %s", domainKey)
				}
			case TypeInt:
				if v, ok := entry.GetIntValue(); ok {
					value, err = NewIntValue(v)
					if err != nil {
						return fmt.Errorf("failed to create int value for domain %s: %w", domainKey, err)
					}
				} else {
					return fmt.Errorf("invalid int value for domain %s", domainKey)
				}
			case TypeFloat:
				if v, ok := entry.GetFloatValue(); ok {
					value, err = NewFloatValue(v)
					if err != nil {
						return fmt.Errorf("failed to create float value for domain %s: %w", domainKey, err)
					}
				} else {
					return fmt.Errorf("invalid float value for domain %s", domainKey)
				}
			default:
				return fmt.Errorf("unsupported type %s for domain %s", entry.Type, domainKey)
			}
		}

		batch.AddCommand(Command{
			Domain: macosDomain,
			Key:    key,
			Value:  value,
		})
	}

	log.Debug("Executing raw defaults batch")
	if err := batch.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute raw defaults: %w", err)
	}

	log.Debug("Raw defaults applied successfully", "count", len(r))
	return nil
}
