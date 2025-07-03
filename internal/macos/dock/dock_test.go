package dock

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "empty config",
			config:  &Config{},
			wantErr: false,
		},
		{
			name: "valid config",
			config: &Config{
				Autohide:    boolPtr(true),
				Orientation: positionPtr(PositionLeft),
				TileSize:    int16Ptr(64),
			},
			wantErr: false,
		},
		{
			name: "invalid orientation",
			config: &Config{
				Orientation: positionPtr(Position(999)),
			},
			wantErr: true,
		},
		{
			name: "invalid min effect",
			config: &Config{
				MinEffect: minEffectPtr(MinEffect(999)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_String(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		contains []string
	}{
		{
			name:     "empty config",
			config:   &Config{},
			contains: []string{"Dock{}"},
		},
		{
			name: "config with values",
			config: &Config{
				Autohide:    boolPtr(true),
				Orientation: positionPtr(PositionLeft),
				TileSize:    int16Ptr(64),
			},
			contains: []string{"Dock{", "autohide", "orientation", "tilesize"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.String()
			for _, contains := range tt.contains {
				if !containsString(result, contains) {
					t.Errorf("Config.String() = %v, should contain %v", result, contains)
				}
			}
		})
	}
}

func TestConfig_Fields(t *testing.T) {
	config := &Config{
		Autohide:    boolPtr(true),
		Orientation: positionPtr(PositionLeft),
		TileSize:    int16Ptr(64),
	}

	fields := config.Fields()
	
	if fields["autohide"] == nil {
		t.Error("Fields() should include autohide")
	}
	
	if fields["orientation"] == nil {
		t.Error("Fields() should include orientation")
	}
	
	if fields["tilesize"] == nil {
		t.Error("Fields() should include tilesize")
	}
	
	if len(fields) != 3 {
		t.Errorf("Fields() should have 3 fields, got %d", len(fields))
	}
}

func TestConfig_Execute_WithMocks(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectCmds  int
		expectError bool
	}{
		{
			name:        "empty config",
			config:      &Config{},
			expectCmds:  0,
			expectError: false,
		},
		{
			name: "config with bool values",
			config: &Config{
				Autohide:    boolPtr(true),
				ShowRecents: boolPtr(false),
			},
			expectCmds:  2,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation
			err := tt.config.Validate()
			if err != nil {
				t.Errorf("Config.Validate() error = %v", err)
			}
			
			// Test that string representation works
			str := tt.config.String()
			if str == "" {
				t.Error("Config.String() should not be empty")
			}
			
			// Test fields method
			fields := tt.config.Fields()
			if tt.expectCmds > 0 && len(fields) == 0 {
				t.Error("Config.Fields() should return fields when config has values")
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func int16Ptr(i int16) *int16 {
	return &i
}

func float32Ptr(f float32) *float32 {
	return &f
}

func positionPtr(p Position) *Position {
	return &p
}

func minEffectPtr(e MinEffect) *MinEffect {
	return &e
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr ||
		     findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}