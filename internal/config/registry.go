package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/RATIU5/fjrd/internal/macos/defaults"
)

type ValueType int

const (
	ValueTypeBool ValueType = iota
	ValueTypeInt
	ValueTypeFloat
	ValueTypeString
	ValueTypeEnum
)

type ConfigField struct {
	Key         string
	Type        ValueType
	Description string
	EnumValues  []string
	Validator   func(any) error
	TOMLTag     string
}

type DomainConfig struct {
	Domain string
	Fields []ConfigField
}

type Registry struct {
	mu      sync.RWMutex
	domains map[string]DomainConfig
}

var globalRegistry = &Registry{
	domains: make(map[string]DomainConfig),
}

func GetRegistry() *Registry {
	return globalRegistry
}

func (r *Registry) RegisterDomain(domain string, config DomainConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	config.Domain = domain
	r.domains[domain] = config
}

func (r *Registry) GetDomain(domain string) (DomainConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, exists := r.domains[domain]
	return config, exists
}

func (r *Registry) ListDomains() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domains := make([]string, 0, len(r.domains))
	for domain := range r.domains {
		domains = append(domains, domain)
	}
	return domains
}

func (r *Registry) ValidateConfig(domain string, config any) error {
	domainConfig, exists := r.GetDomain(domain)
	if !exists {
		return fmt.Errorf("unknown domain: %s", domain)
	}

	configValue := reflect.ValueOf(config)
	if configValue.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
	}

	if configValue.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a struct, got %T", config)
	}

	configType := configValue.Type()

	for _, field := range domainConfig.Fields {
		structField, found := configType.FieldByName(field.Key)
		if !found {
			continue
		}

		fieldValue := configValue.FieldByName(field.Key)
		if !fieldValue.IsValid() || fieldValue.IsZero() {
			continue
		}

		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				continue
			}
			fieldValue = fieldValue.Elem()
		}

		if err := r.validateFieldValue(field, fieldValue.Interface(), structField); err != nil {
			return fmt.Errorf("validation failed for field %s: %w", field.Key, err)
		}
	}

	return nil
}

func (r *Registry) validateFieldValue(field ConfigField, value any, structField reflect.StructField) error {
	switch field.Type {
	case ValueTypeBool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case ValueTypeInt:
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case ValueTypeFloat:
		switch value.(type) {
		case float32, float64:
		default:
			return fmt.Errorf("expected float, got %T", value)
		}
	case ValueTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case ValueTypeEnum:
		strValue := fmt.Sprintf("%v", value)
		for _, enumValue := range field.EnumValues {
			if strValue == enumValue {
				goto validEnum
			}
		}
		return fmt.Errorf("invalid enum value %q, must be one of: %s",
			strValue, strings.Join(field.EnumValues, ", "))
	validEnum:
	}

	if field.Validator != nil {
		if err := field.Validator(value); err != nil {
			return err
		}
	}

	return nil
}

func (r *Registry) GenerateDefaultsCommands(domain string, config any) ([]defaults.Command, error) {
	domainConfig, exists := r.GetDomain(domain)
	if !exists {
		return nil, fmt.Errorf("unknown domain: %s", domain)
	}

	configValue := reflect.ValueOf(config)
	if configValue.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
	}

	if configValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("config must be a struct, got %T", config)
	}

	commands := make([]defaults.Command, 0)

	for _, field := range domainConfig.Fields {
		fieldValue := configValue.FieldByName(field.Key)
		if !fieldValue.IsValid() || fieldValue.IsZero() {
			continue
		}

		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				continue
			}
			fieldValue = fieldValue.Elem()
		}

		var value defaults.Value
		var err error

		switch field.Type {
		case ValueTypeBool:
			value = defaults.NewBoolValue(fieldValue.Bool())
		case ValueTypeInt:
			value, err = defaults.NewIntValue(fieldValue.Int())
		case ValueTypeFloat:
			value, err = defaults.NewFloatValue(fieldValue.Float())
		case ValueTypeString:
			value = defaults.NewStringValue(fieldValue.String())
		case ValueTypeEnum:
			strValue := fmt.Sprintf("%v", fieldValue.Interface())
			value = defaults.NewEnumValue(strValue, field.EnumValues)
		default:
			return nil, fmt.Errorf("unsupported field type: %v", field.Type)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create value for field %s: %w", field.Key, err)
		}

		commands = append(commands, defaults.Command{
			Domain: domain,
			Key:    getTOMLKey(field),
			Value:  value,
		})
	}

	return commands, nil
}

func getTOMLKey(field ConfigField) string {
	if field.TOMLTag != "" {
		return field.TOMLTag
	}

	return strings.ToLower(strings.ReplaceAll(field.Key, "_", "-"))
}

func RegisterFromStruct(domain string, structType any, description string) error {
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct type, got %T", structType)
	}

	fields := make([]ConfigField, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		tomlTag := field.Tag.Get("toml")
		if tomlTag == "-" {
			continue
		}

		configField := ConfigField{
			Key:     field.Name,
			TOMLTag: extractTOMLKey(tomlTag),
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		switch fieldType.Kind() {
		case reflect.Bool:
			configField.Type = ValueTypeBool
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			configField.Type = ValueTypeInt
		case reflect.Float32, reflect.Float64:
			configField.Type = ValueTypeFloat
		case reflect.String:
			if fieldType.Name() != "string" {
				configField.Type = ValueTypeEnum
			} else {
				configField.Type = ValueTypeString
			}
		default:
			if fieldType.Kind() == reflect.String {
				configField.Type = ValueTypeEnum
			} else {
				continue
			}
		}

		fields = append(fields, configField)
	}

	config := DomainConfig{
		Domain: domain,
		Fields: fields,
	}

	globalRegistry.RegisterDomain(domain, config)
	return nil
}

func extractTOMLKey(tag string) string {
	if tag == "" {
		return ""
	}

	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return ""
	}

	return strings.TrimSpace(parts[0])
}
