package errors

import (
	"errors"
	"fmt"
	"strings"
)

type ConfigurationError struct {
	Component string
	Operation string
	Field     string
	Value     any
	Err       error
}

func (e *ConfigurationError) Error() string {
	var parts []string

	if e.Component != "" {
		parts = append(parts, fmt.Sprintf("component: %s", e.Component))
	}

	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("operation: %s", e.Operation))
	}

	if e.Field != "" {
		parts = append(parts, fmt.Sprintf("field: %s", e.Field))
	}

	if e.Value != nil {
		parts = append(parts, fmt.Sprintf("value: %v", e.Value))
	}

	context := strings.Join(parts, ", ")
	if context != "" {
		return fmt.Sprintf("configuration error (%s): %v", context, e.Err)
	}

	return fmt.Sprintf("configuration error: %v", e.Err)
}

func (e *ConfigurationError) Unwrap() error {
	return e.Err
}

func NewConfigurationError(component, operation, field string, value any, err error) *ConfigurationError {
	return &ConfigurationError{
		Component: component,
		Operation: operation,
		Field:     field,
		Value:     value,
		Err:       err,
	}
}

type ValidationError struct {
	Field    string
	Value    any
	Expected string
	Err      error
}

func (e *ValidationError) Error() string {
	if e.Expected != "" {
		return fmt.Sprintf("validation error for field %s (value: %v, expected: %s): %v",
			e.Field, e.Value, e.Expected, e.Err)
	}
	return fmt.Sprintf("validation error for field %s (value: %v): %v",
		e.Field, e.Value, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func NewValidationError(field string, value any, expected string, err error) *ValidationError {
	return &ValidationError{
		Field:    field,
		Value:    value,
		Expected: expected,
		Err:      err,
	}
}

type ExecutionError struct {
	Command   string
	Arguments []string
	Err       error
}

func (e *ExecutionError) Error() string {
	if len(e.Arguments) > 0 {
		return fmt.Sprintf("execution error for command '%s %s': %v",
			e.Command, strings.Join(e.Arguments, " "), e.Err)
	}
	return fmt.Sprintf("execution error for command '%s': %v", e.Command, e.Err)
}

func (e *ExecutionError) Unwrap() error {
	return e.Err
}

func NewExecutionError(command string, arguments []string, err error) *ExecutionError {
	return &ExecutionError{
		Command:   command,
		Arguments: arguments,
		Err:       err,
	}
}

type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	
	var messages []string
	for i, err := range e.Errors {
		messages = append(messages, fmt.Sprintf("[%d] %s", i+1, err.Error()))
	}
	
	return fmt.Sprintf("multiple errors occurred: %s", strings.Join(messages, "; "))
}

func (e *MultiError) Unwrap() []error {
	return e.Errors
}

func (e *MultiError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

func (e *MultiError) ToError() error {
	if !e.HasErrors() {
		return nil
	}
	return e
}

func NewMultiError(errs ...error) *MultiError {
	me := &MultiError{}
	for _, err := range errs {
		me.Add(err)
	}
	return me
}

func Combine(errs ...error) error {
	me := NewMultiError(errs...)
	return me.ToError()
}

func IsConfigurationError(err error) bool {
	var configErr *ConfigurationError
	return errors.As(err, &configErr)
}

func IsValidationError(err error) bool {
	var validErr *ValidationError
	return errors.As(err, &validErr)
}

func IsExecutionError(err error) bool {
	var execErr *ExecutionError
	return errors.As(err, &execErr)
}

func WrapConfigError(component, operation, field string, value any, err error) error {
	if err == nil {
		return nil
	}
	return NewConfigurationError(component, operation, field, value, err)
}

func WrapValidationError(field string, value any, expected string, err error) error {
	if err == nil {
		return nil
	}
	return NewValidationError(field, value, expected, err)
}

func WrapExecutionError(command string, arguments []string, err error) error {
	if err == nil {
		return nil
	}
	return NewExecutionError(command, arguments, err)
}