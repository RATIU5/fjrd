package config

import (
	"fmt"
	"reflect"
	"slices"
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

func IndentString(s string, indent string) string {
	if s == "" {
		return s
	}

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i > 0 && i < len(lines)-1 && line != "" {
			lines[i] = indent + line
		}
	}

	return strings.Join(lines, "\n")
}

func MergeValidationErrors(errs ...error) error {
	var validErrs []error
	for _, err := range errs {
		if err != nil {
			validErrs = append(validErrs, err)
		}
	}

	if len(validErrs) == 0 {
		return nil
	}

	if len(validErrs) == 1 {
		return validErrs[0]
	}

	var messages []string
	for _, err := range validErrs {
		messages = append(messages, err.Error())
	}

	return fmt.Errorf("multiple validation errors: %s", strings.Join(messages, "; "))
}

type FieldValidator struct {
	FieldName string
	Value     any
	Rules     []ValidationRule
}

type ValidationRule func(any) error

func (fv FieldValidator) Validate() error {
	for _, rule := range fv.Rules {
		if err := rule(fv.Value); err != nil {
			return fmt.Errorf("validation failed for field %s: %w", fv.FieldName, err)
		}
	}
	return nil
}

func NotNil() ValidationRule {
	return func(value any) error {
		if value == nil {
			return fmt.Errorf("value cannot be nil")
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr && rv.IsNil() {
			return fmt.Errorf("value cannot be nil")
		}

		return nil
	}
}

func InRange[T comparable](min, max T) ValidationRule {
	return func(value any) error {
		v, ok := value.(T)
		if !ok {
			return fmt.Errorf("expected type %T, got %T", min, value)
		}

		switch any(min).(type) {
		case int, int8, int16, int32, int64:
			minInt := reflect.ValueOf(min).Int()
			maxInt := reflect.ValueOf(max).Int()
			vInt := reflect.ValueOf(v).Int()
			if vInt < minInt || vInt > maxInt {
				return fmt.Errorf("value %v is not in range [%v, %v]", v, min, max)
			}
		case uint, uint8, uint16, uint32, uint64:
			minUint := reflect.ValueOf(min).Uint()
			maxUint := reflect.ValueOf(max).Uint()
			vUint := reflect.ValueOf(v).Uint()
			if vUint < minUint || vUint > maxUint {
				return fmt.Errorf("value %v is not in range [%v, %v]", v, min, max)
			}
		case float32, float64:
			minFloat := reflect.ValueOf(min).Float()
			maxFloat := reflect.ValueOf(max).Float()
			vFloat := reflect.ValueOf(v).Float()
			if vFloat < minFloat || vFloat > maxFloat {
				return fmt.Errorf("value %v is not in range [%v, %v]", v, min, max)
			}
		default:
			return fmt.Errorf("range validation not supported for type %T", min)
		}

		return nil
	}
}

func OneOf[T comparable](values ...T) ValidationRule {
	return func(value any) error {
		v, ok := value.(T)
		if !ok {
			return fmt.Errorf("expected type %T, got %T", values[0], value)
		}

		if slices.Contains(values, v) {
			return nil
		}

		return fmt.Errorf("value %v is not one of %v", v, values)
	}
}
