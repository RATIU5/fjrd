package menubar

import (
	"context"

	"github.com/RATIU5/fjrd/internal/errors"
	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/shared"
)

type Config struct {
	ClockFlashDateSeparators *bool   `toml:"clock-flash-date-separators,omitmepty"`
	ClockDateFormat          *string `toml:"clock-date-format,omitempty"`
}

func (m *Config) Validate() error {
	return nil
}

func (m *Config) String() string {
	return shared.FormatConfig("Menubar", m)
}

func (m *Config) Fields() map[string]any {
	fields := make(map[string]any)

	if m.ClockFlashDateSeparators != nil {
		fields["clock-flash-date-separators"] = m.ClockFlashDateSeparators
	}
	if m.ClockDateFormat != nil {
		fields["clock-date-format"] = m.ClockDateFormat
	}

	return fields
}

func (m *Config) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("menubar")
	log.Debug("Configuring menubar settings")

	batch := defaults.NewBatchExecutor()
	const clockDomain = "com.apple.menuextra.clock"

	if m.ClockFlashDateSeparators != nil {
		batch.AddBool(clockDomain, "FlashDateSeparators", *m.ClockFlashDateSeparators)
	}

	if m.ClockDateFormat != nil {
		batch.AddString(clockDomain, "DateFormat", *m.ClockDateFormat)
	}

	log.Debug("Applying menubar defaults")
	if err := batch.Execute(ctx, log); err != nil {
		return errors.WrapConfigError("menubar", "execute_batch", "", nil, err)
	}

	log.Debug("Menubar configuration applied successfully")
	return nil
}
