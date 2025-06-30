package defaults

import (
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
	RawValue any          `toml:"value" comment:"defaults key"`
	Type     DefaultsType `toml:"type" comment:"type of defaults domain"`
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
		if v, ok := e.RawValue.(bool); ok {
			return v, true
		}
	}
	return false, false
}

func (e *RawEntry) GetIntValue() (int16, bool) {
	if e.Type == TypeInt {
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

func (e *RawEntry) Validate() error {
	switch e.Type {
	case TypeString:
		if value, isString := e.GetStringValue(); !isString {
			return fmt.Errorf("%v is not of expected type string", value)
		}
	case TypeInt:
		if value, isInt := e.GetIntValue(); !isInt {
			return fmt.Errorf("%v is not of expected type int", value)
		}
	case TypeFloat:
		if value, isFloat := e.GetFloatValue(); !isFloat {
			return fmt.Errorf("%v is not of expected type float", value)
		}
	case TypeBool:
		if value, isBool := e.GetBoolValue(); !isBool {
			return fmt.Errorf("%v is not of expected type bool", value)
		}
	}
	return nil
}

type Raw map[string]RawEntry

func (r *Raw) String() string {
	if len(*r) == 0 {
		return "DefaultsRaw{}"
	}

	var parts []string
	for key, entry := range *r {
		parts = append(parts, fmt.Sprintf("  %q: %s", key, entry.String()))
	}
	return fmt.Sprintf("DefaultsRaw{\n%s\n}", strings.Join(parts, "\n"))
}

func (r *Raw) Validate() error {
	for _, value := range *r {
		if err := value.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Raw) Execute() error {
	batch := NewBatchExecutor()

	for domainKey, entry := range *r {
		// Split domain into domain and key (assuming format "domain.key")
		domainParts := strings.Split(domainKey, ".")
		if len(domainParts) < 2 {
			return fmt.Errorf("invalid domain format %s, expected format: com.apple.domain.key", domainKey)
		}
		
		// Extract the key (last part) and domain (everything else)
		key := domainParts[len(domainParts)-1]
		macosDomain := strings.Join(domainParts[:len(domainParts)-1], ".")

		// Create the appropriate value type
		var value Value
		var err error

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

		batch.AddCommand(Command{
			Domain: macosDomain,
			Key:    key,
			Value:  value,
		})
	}

	if err := batch.Execute(); err != nil {
		return fmt.Errorf("failed to execute raw defaults: %w", err)
	}

	return nil
}
