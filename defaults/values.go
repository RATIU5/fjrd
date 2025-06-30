package defaults

import (
	"fmt"
	"strconv"
)

// Value represents a typed value that can be converted to defaults command format
type Value interface {
	Type() ValueType
	String() string
	Validate() error
}

// BoolValue represents a boolean value
type BoolValue struct {
	Value bool
}

// NewBoolValue creates a new boolean value
func NewBoolValue(value bool) *BoolValue {
	return &BoolValue{Value: value}
}

// Type returns the value type
func (b *BoolValue) Type() ValueType {
	return BoolType
}

// String returns the string representation
func (b *BoolValue) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// Validate validates the boolean value
func (b *BoolValue) Validate() error {
	return nil // Boolean values are always valid
}

// StringValue represents a string value
type StringValue struct {
	Value string
}

// NewStringValue creates a new string value
func NewStringValue(value string) *StringValue {
	return &StringValue{Value: value}
}

// Type returns the value type
func (s *StringValue) Type() ValueType {
	return StringType
}

// String returns the string representation
func (s *StringValue) String() string {
	return s.Value
}

// Validate validates the string value
func (s *StringValue) Validate() error {
	if s.Value == "" {
		return fmt.Errorf("string value cannot be empty")
	}
	return nil
}

// IntValue represents an integer value
type IntValue struct {
	Value int64
}

// NewIntValue creates a new integer value
func NewIntValue(value interface{}) (*IntValue, error) {
	var intVal int64
	
	switch v := value.(type) {
	case int:
		intVal = int64(v)
	case int8:
		intVal = int64(v)
	case int16:
		intVal = int64(v)
	case int32:
		intVal = int64(v)
	case int64:
		intVal = v
	case uint:
		intVal = int64(v)
	case uint8:
		intVal = int64(v)
	case uint16:
		intVal = int64(v)
	case uint32:
		intVal = int64(v)
	case uint64:
		if v > 9223372036854775807 { // max int64
			return nil, fmt.Errorf("value %d is too large for int64", v)
		}
		intVal = int64(v)
	default:
		return nil, fmt.Errorf("cannot convert %T to int", value)
	}
	
	return &IntValue{Value: intVal}, nil
}

// Type returns the value type
func (i *IntValue) Type() ValueType {
	return IntType
}

// String returns the string representation
func (i *IntValue) String() string {
	return strconv.FormatInt(i.Value, 10)
}

// Validate validates the integer value
func (i *IntValue) Validate() error {
	return nil // All int64 values are valid
}

// FloatValue represents a floating-point value
type FloatValue struct {
	Value    float64
	Precision int
}

// NewFloatValue creates a new float value with default precision
func NewFloatValue(value interface{}) (*FloatValue, error) {
	return NewFloatValueWithPrecision(value, 2)
}

// NewFloatValueWithPrecision creates a new float value with specified precision
func NewFloatValueWithPrecision(value interface{}, precision int) (*FloatValue, error) {
	var floatVal float64
	
	switch v := value.(type) {
	case float32:
		floatVal = float64(v)
	case float64:
		floatVal = v
	case int:
		floatVal = float64(v)
	case int8:
		floatVal = float64(v)
	case int16:
		floatVal = float64(v)
	case int32:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	default:
		return nil, fmt.Errorf("cannot convert %T to float", value)
	}
	
	if precision < 0 {
		precision = 2
	}
	
	return &FloatValue{Value: floatVal, Precision: precision}, nil
}

// Type returns the value type
func (f *FloatValue) Type() ValueType {
	return FloatType
}

// String returns the string representation
func (f *FloatValue) String() string {
	return strconv.FormatFloat(f.Value, 'f', f.Precision, 64)
}

// Validate validates the float value
func (f *FloatValue) Validate() error {
	return nil // All float64 values are valid
}

// EnumValue represents an enumerated string value with validation
type EnumValue struct {
	Value         string
	AllowedValues []string
}

// NewEnumValue creates a new enum value
func NewEnumValue(value string, allowedValues []string) *EnumValue {
	return &EnumValue{
		Value:         value,
		AllowedValues: allowedValues,
	}
}

// Type returns the value type
func (e *EnumValue) Type() ValueType {
	return StringType
}

// String returns the string representation
func (e *EnumValue) String() string {
	return e.Value
}

// Validate validates the enum value
func (e *EnumValue) Validate() error {
	for _, allowed := range e.AllowedValues {
		if e.Value == allowed {
			return nil
		}
	}
	return fmt.Errorf("value %q is not allowed, must be one of: %v", e.Value, e.AllowedValues)
}

// ValueFactory helps create values from interface{} types
type ValueFactory struct{}

// NewValueFactory creates a new value factory
func NewValueFactory() *ValueFactory {
	return &ValueFactory{}
}

// CreateValue creates a Value from an interface{} and type specification
func (f *ValueFactory) CreateValue(value interface{}, valueType ValueType) (Value, error) {
	switch valueType {
	case BoolType:
		if b, ok := value.(bool); ok {
			return NewBoolValue(b), nil
		}
		return nil, fmt.Errorf("expected bool, got %T", value)
	
	case StringType:
		if s, ok := value.(string); ok {
			return NewStringValue(s), nil
		}
		return nil, fmt.Errorf("expected string, got %T", value)
	
	case IntType:
		return NewIntValue(value)
	
	case FloatType:
		return NewFloatValue(value)
	
	default:
		return nil, fmt.Errorf("unsupported value type: %s", valueType)
	}
}