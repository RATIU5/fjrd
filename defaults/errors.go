package defaults

import (
	"fmt"
	"strings"
)

// ConfigurationError represents an error that occurred during configuration
type ConfigurationError struct {
	Component string
	Operation string
	Field     string
	Value     any
	Err       error
}

// Error implements the error interface
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

// Unwrap returns the underlying error
func (e *ConfigurationError) Unwrap() error {
	return e.Err
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(component, operation, field string, value interface{}, err error) *ConfigurationError {
	return &ConfigurationError{
		Component: component,
		Operation: operation,
		Field:     field,
		Value:     value,
		Err:       err,
	}
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field    string
	Value    interface{}
	Expected string
	Err      error
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Expected != "" {
		return fmt.Sprintf("validation error for field %s (value: %v, expected: %s): %v",
			e.Field, e.Value, e.Expected, e.Err)
	}
	return fmt.Sprintf("validation error for field %s (value: %v): %v",
		e.Field, e.Value, e.Err)
}

// Unwrap returns the underlying error
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, expected string, err error) *ValidationError {
	return &ValidationError{
		Field:    field,
		Value:    value,
		Expected: expected,
		Err:      err,
	}
}

// ExecutionError represents an execution failure
type ExecutionError struct {
	Command   string
	Arguments []string
	Err       error
}

// Error implements the error interface
func (e *ExecutionError) Error() string {
	if len(e.Arguments) > 0 {
		return fmt.Sprintf("execution error for command '%s %s': %v",
			e.Command, strings.Join(e.Arguments, " "), e.Err)
	}
	return fmt.Sprintf("execution error for command '%s': %v", e.Command, e.Err)
}

// Unwrap returns the underlying error
func (e *ExecutionError) Unwrap() error {
	return e.Err
}

// NewExecutionError creates a new execution error
func NewExecutionError(command string, arguments []string, err error) *ExecutionError {
	return &ExecutionError{
		Command:   command,
		Arguments: arguments,
		Err:       err,
	}
}
