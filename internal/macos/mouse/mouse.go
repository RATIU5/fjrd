package mouse

import (
	"context"

	"github.com/RATIU5/fjrd/internal/errors"
	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/shared"
)

// TODO: Need to specify that these cfg properties need a restart

type Config struct {
	Acceleration *bool    `toml:"acceleration,omitempty"`
	Speed        *float32 `toml:"speed,omitempty"`
}

func (m *Config) Validate() error {
	return nil
}

func (m *Config) String() string {
	return shared.FormatConfig("Mouse", m)
}

func (m *Config) Fields() map[string]any {
	fields := make(map[string]any)

	if m.Acceleration != nil {
		fields["acceleration"] = !*m.Acceleration
	}
	if m.Speed != nil {
		fields["speed"] = m.Speed
	}

	return fields
}

func (m *Config) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("mouse")
	log.Debug("Configuring mouse settings")

	multiErr := errors.NewMultiError()

	batch := defaults.NewBatchExecutor()
	const nsDomain = "NSGlobalDomain"

	if m.Acceleration != nil {
		batch.AddBool(nsDomain, "com.apple.mouse.linear", *m.Acceleration)
	}

	if m.Speed != nil {
		if err := batch.AddFloat(nsDomain, "com.apple.mouse.scaling", *m.Speed); err != nil {
			multiErr.Add(errors.WrapConfigError("mouse", "add_command", "speed", *m.Speed, err))
		}
	}

	log.Debug("Applying mouse defaults")
	if err := batch.Execute(ctx, log); err != nil {
		return errors.WrapConfigError("mouse", "execute_batch", "", nil, err)
	}

	log.Debug("Mouse configuration applied successfully")
	return nil
}
