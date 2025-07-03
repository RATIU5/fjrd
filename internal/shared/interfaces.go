package shared

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type Validator interface {
	Validate() error
}

type Stringer interface {
	String() string
}

type ConfigFields interface {
	Fields() map[string]any
}

type CompositeValidator []Validator

func (cv CompositeValidator) Validate() error {
	for _, v := range cv {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func ValidateAll(validators ...Validator) error {
	return CompositeValidator(validators).Validate()
}

func FormatConfig(name string, config any) string {
	if cf, ok := config.(ConfigFields); ok {
		return formatConfigFromFields(name, cf.Fields())
	}
	
	return formatConfigFromReflection(name, config)
}

func formatConfigFromFields(name string, fields map[string]any) string {
	if len(fields) == 0 {
		return fmt.Sprintf("%s{}", name)
	}
	
	var parts []string
	var keys []string
	
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	
	for _, key := range keys {
		value := fields[key]
		if value == nil {
			continue
		}
		
		var formattedValue string
		switch v := value.(type) {
		case *bool:
			if v != nil {
				formattedValue = fmt.Sprintf("%t", *v)
			} else {
				continue
			}
		case *int, *int8, *int16, *int32, *int64:
			if v != nil {
				formattedValue = fmt.Sprintf("%d", reflect.ValueOf(v).Elem().Int())
			} else {
				continue
			}
		case *uint, *uint8, *uint16, *uint32, *uint64:
			if v != nil {
				formattedValue = fmt.Sprintf("%d", reflect.ValueOf(v).Elem().Uint())
			} else {
				continue
			}
		case *float32, *float64:
			if v != nil {
				formattedValue = fmt.Sprintf("%.2f", reflect.ValueOf(v).Elem().Float())
			} else {
				continue
			}
		case *string:
			if v != nil {
				formattedValue = *v
			} else {
				continue
			}
		case fmt.Stringer:
			formattedValue = v.String()
		case string:
			formattedValue = v
		case bool:
			formattedValue = fmt.Sprintf("%t", v)
		case int, int8, int16, int32, int64:
			formattedValue = fmt.Sprintf("%d", v)
		case uint, uint8, uint16, uint32, uint64:
			formattedValue = fmt.Sprintf("%d", v)
		case float32, float64:
			formattedValue = fmt.Sprintf("%.2f", v)
		default:
			if reflect.ValueOf(v).Kind() == reflect.Ptr {
				rv := reflect.ValueOf(v)
				if rv.IsNil() {
					continue
				}
				elem := rv.Elem()
				if elem.CanInterface() {
					formattedValue = fmt.Sprintf("%v", elem.Interface())
				} else {
					continue
				}
			} else {
				formattedValue = fmt.Sprintf("%v", v)
			}
		}
		
		parts = append(parts, fmt.Sprintf("%s: %s", key, formattedValue))
	}
	
	if len(parts) == 0 {
		return fmt.Sprintf("%s{}", name)
	}
	
	return fmt.Sprintf("%s{%s}", name, strings.Join(parts, ", "))
}

func formatConfigFromReflection(name string, config any) string {
	value := reflect.ValueOf(config)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return fmt.Sprintf("%s{}", name)
		}
		value = value.Elem()
	}
	
	if value.Kind() != reflect.Struct {
		return fmt.Sprintf("%s{%v}", name, config)
	}
	
	typ := value.Type()
	fields := make(map[string]any)
	
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		structField := typ.Field(i)
		
		if !structField.IsExported() {
			continue
		}
		
		if field.IsZero() {
			continue
		}
		
		key := strings.ToLower(structField.Name)
		if tomlTag := structField.Tag.Get("toml"); tomlTag != "" && tomlTag != "-" {
			parts := strings.Split(tomlTag, ",")
			if len(parts) > 0 && parts[0] != "" {
				key = parts[0]
			}
		}
		
		if field.CanInterface() {
			fields[key] = field.Interface()
		}
	}
	
	return formatConfigFromFields(name, fields)
}