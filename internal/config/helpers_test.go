package config

import (
	"testing"
)

func TestFormatConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   any
		expected string
	}{
		{
			name:     "empty config",
			config:   &struct{}{},
			expected: "TestConfig{}",
		},
		{
			name: "config with fields",
			config: &struct {
				Name  string `toml:"name"`
				Value int    `toml:"value"`
			}{
				Name:  "test",
				Value: 42,
			},
			expected: "TestConfig{name: test, value: 42}",
		},
		{
			name: "config with pointer fields",
			config: &struct {
				Enabled *bool `toml:"enabled"`
				Count   *int  `toml:"count"`
			}{
				Enabled: boolPtr(true),
				Count:   intPtr(10),
			},
			expected: "TestConfig{count: 10, enabled: true}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatConfig("TestConfig", tt.config)
			if result != tt.expected {
				t.Errorf("FormatConfig() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCompositeValidator(t *testing.T) {
	tests := []struct {
		name       string
		validators []Validator
		wantErr    bool
	}{
		{
			name:       "empty validators",
			validators: []Validator{},
			wantErr:    false,
		},
		{
			name: "all valid",
			validators: []Validator{
				&mockValidator{err: nil},
				&mockValidator{err: nil},
			},
			wantErr: false,
		},
		{
			name: "one invalid",
			validators: []Validator{
				&mockValidator{err: nil},
				&mockValidator{err: &validationError{"test error"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cv := CompositeValidator(tt.validators)
			err := cv.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CompositeValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFieldValidator(t *testing.T) {
	tests := []struct {
		name      string
		validator FieldValidator
		wantErr   bool
	}{
		{
			name: "no rules",
			validator: FieldValidator{
				FieldName: "test",
				Value:     "value",
				Rules:     []ValidationRule{},
			},
			wantErr: false,
		},
		{
			name: "passing rule",
			validator: FieldValidator{
				FieldName: "test",
				Value:     "value",
				Rules: []ValidationRule{
					func(any) error { return nil },
				},
			},
			wantErr: false,
		},
		{
			name: "failing rule",
			validator: FieldValidator{
				FieldName: "test",
				Value:     "value",
				Rules: []ValidationRule{
					func(any) error { return &validationError{"test error"} },
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("FieldValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInRange(t *testing.T) {
	tests := []struct {
		name    string
		rule    ValidationRule
		value   any
		wantErr bool
	}{
		{
			name:    "int in range",
			rule:    InRange(1, 10),
			value:   5,
			wantErr: false,
		},
		{
			name:    "int below range",
			rule:    InRange(1, 10),
			value:   0,
			wantErr: true,
		},
		{
			name:    "int above range",
			rule:    InRange(1, 10),
			value:   11,
			wantErr: true,
		},
		{
			name:    "float in range",
			rule:    InRange(1.0, 10.0),
			value:   5.5,
			wantErr: false,
		},
		{
			name:    "wrong type",
			rule:    InRange(1, 10),
			value:   "string",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("InRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOneOf(t *testing.T) {
	tests := []struct {
		name    string
		rule    ValidationRule
		value   any
		wantErr bool
	}{
		{
			name:    "value in list",
			rule:    OneOf("a", "b", "c"),
			value:   "b",
			wantErr: false,
		},
		{
			name:    "value not in list",
			rule:    OneOf("a", "b", "c"),
			value:   "d",
			wantErr: true,
		},
		{
			name:    "wrong type",
			rule:    OneOf("a", "b", "c"),
			value:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("OneOf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNotNil(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{
			name:    "non-nil value",
			value:   "test",
			wantErr: false,
		},
		{
			name:    "nil value",
			value:   nil,
			wantErr: true,
		},
		{
			name:    "nil pointer",
			value:   (*string)(nil),
			wantErr: true,
		},
		{
			name:    "non-nil pointer",
			value:   stringPtr("test"),
			wantErr: false,
		},
	}

	rule := NotNil()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NotNil() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockValidator struct {
	err error
}

func (m *mockValidator) Validate() error {
	return m.err
}

type validationError struct {
	msg string
}

func (e *validationError) Error() string {
	return e.msg
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
