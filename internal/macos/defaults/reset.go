package defaults

import "fmt"

type ResetStringValue struct{}

func NewResetStringValue() *ResetStringValue {
	return &ResetStringValue{}
}

func (r *ResetStringValue) Type() ValueType {
	return StringType
}

func (r *ResetStringValue) String() string {
	return "default"
}

func (r *ResetStringValue) IsReset() bool {
	return true
}

func (r *ResetStringValue) Validate() error {
	return nil
}

type ResetBoolValue struct{}

func NewResetBoolValue() *ResetBoolValue {
	return &ResetBoolValue{}
}

func (r *ResetBoolValue) Type() ValueType {
	return BoolType
}

func (r *ResetBoolValue) String() string {
	return "default"
}

func (r *ResetBoolValue) IsReset() bool {
	return true
}

func (r *ResetBoolValue) Validate() error {
	return nil
}

type ResetIntValue struct{}

func NewResetIntValue() *ResetIntValue {
	return &ResetIntValue{}
}

func (r *ResetIntValue) Type() ValueType {
	return IntType
}

func (r *ResetIntValue) String() string {
	return "default"
}

func (r *ResetIntValue) IsReset() bool {
	return true
}

func (r *ResetIntValue) Validate() error {
	return nil
}

type ResetFloatValue struct{}

func NewResetFloatValue() *ResetFloatValue {
	return &ResetFloatValue{}
}

func (r *ResetFloatValue) Type() ValueType {
	return FloatType
}

func (r *ResetFloatValue) String() string {
	return "default"
}

func (r *ResetFloatValue) IsReset() bool {
	return true
}

func (r *ResetFloatValue) Validate() error {
	return nil
}

func NewValueOrReset(value interface{}, valueType ValueType, isNull bool) (Value, error) {
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

	factory := NewValueFactory()
	return factory.CreateValue(value, valueType)
}