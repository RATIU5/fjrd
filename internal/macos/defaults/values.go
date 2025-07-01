package defaults

import (
	"fmt"
	"strconv"
)

type ValueType string

const (
	BoolType   ValueType = "-bool"
	StringType ValueType = "-string"
	IntType    ValueType = "-int"
	FloatType  ValueType = "-float"
)

type Value interface {
	Type() ValueType
	String() string
	Validate() error
}

type ResetValue interface {
	IsReset() bool
}

type BoolValue struct {
	Value bool
}

func NewBoolValue(value bool) *BoolValue {
	return &BoolValue{Value: value}
}

func (b *BoolValue) Type() ValueType {
	return BoolType
}

func (b *BoolValue) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

func (b *BoolValue) Validate() error {
	return nil
}

type StringValue struct {
	Value string
	Reset bool
}

func NewStringValue(value string) *StringValue {
	if value == "default" {
		return &StringValue{Value: "", Reset: true}
	}
	return &StringValue{Value: value, Reset: false}
}

func (s *StringValue) Type() ValueType {
	return StringType
}

func (s *StringValue) String() string {
	if s.Reset {
		return "default"
	}
	return s.Value
}

func (s *StringValue) IsReset() bool {
	return s.Reset
}

func (s *StringValue) Validate() error {
	if !s.Reset && s.Value == "" {
		return fmt.Errorf("string value cannot be empty")
	}
	return nil
}

type IntValue struct {
	Value int64
}

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
		if v > 9223372036854775807 {
			return nil, fmt.Errorf("value %d is too large for int64", v)
		}
		intVal = int64(v)
	default:
		return nil, fmt.Errorf("cannot convert %T to int", value)
	}
	
	return &IntValue{Value: intVal}, nil
}

func (i *IntValue) Type() ValueType {
	return IntType
}

func (i *IntValue) String() string {
	return strconv.FormatInt(i.Value, 10)
}

func (i *IntValue) Validate() error {
	return nil
}

type FloatValue struct {
	Value    float64
	Precision int
}

func NewFloatValue(value interface{}) (*FloatValue, error) {
	return NewFloatValueWithPrecision(value, 2)
}

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

func (f *FloatValue) Type() ValueType {
	return FloatType
}

func (f *FloatValue) String() string {
	return strconv.FormatFloat(f.Value, 'f', f.Precision, 64)
}

func (f *FloatValue) Validate() error {
	return nil
}

type EnumValue struct {
	Value         string
	AllowedValues []string
}

func NewEnumValue(value string, allowedValues []string) *EnumValue {
	return &EnumValue{
		Value:         value,
		AllowedValues: allowedValues,
	}
}

func (e *EnumValue) Type() ValueType {
	return StringType
}

func (e *EnumValue) String() string {
	return e.Value
}

func (e *EnumValue) Validate() error {
	for _, allowed := range e.AllowedValues {
		if e.Value == allowed {
			return nil
		}
	}
	return fmt.Errorf("value %q is not allowed, must be one of: %v", e.Value, e.AllowedValues)
}

type ValueFactory struct{}

func NewValueFactory() *ValueFactory {
	return &ValueFactory{}
}

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

func (f *ValueFactory) CreateValueWithReset(value interface{}, valueType ValueType, isNull bool) (Value, error) {
	if isNull {
		switch valueType {
		case BoolType:
			return NewResetBoolValue(), nil
		case StringType:
			return NewResetStringValue(), nil
		case IntType:
			return NewResetIntValue(), nil
		case FloatType:
			return NewResetFloatValue(), nil
		default:
			return nil, fmt.Errorf("unsupported value type for reset: %s", valueType)
		}
	}

	return f.CreateValue(value, valueType)
}