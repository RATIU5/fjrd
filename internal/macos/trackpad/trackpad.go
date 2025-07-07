package trackpad

import (
	"context"
	"fmt"

	"github.com/RATIU5/fjrd/internal/errors"
	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/shared"
)

type Config struct {
	ClickWeight     *int16 `toml:"click-weight,omitempty"`
	ThreeFingerDrag *bool  `toml:"three-finger-drag,omitempty"`
}

func (t *Config) Validate() error {
	if *t.ClickWeight < 0 || *t.ClickWeight > 3 {
		return fmt.Errorf("invalid click-weight value: %d", *t.ClickWeight)
	}
	return nil
}

func (t *Config) String() string {
	return shared.FormatConfig("Trackpad", t)
}

func (t *Config) Fields() map[string]any {
	fields := make(map[string]any)

	if t.ClickWeight != nil {
		fields["click-weight"] = t.ClickWeight
	}
	if t.ThreeFingerDrag != nil {
		fields["three-finger-drag"] = t.ThreeFingerDrag
	}

	return fields
}

func (t *Config) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("trackpad")
	log.Debug("Configuring trackpad settings")

	batch := defaults.NewBatchExecutor()
	const trackpadDomain = "com.apple.AppleMultitouchTrackpad"

	multiErr := errors.NewMultiError()

	if t.ClickWeight != nil {
		if err := batch.AddInt(trackpadDomain, "FirstClickThreshold", *t.ClickWeight); err != nil {
			multiErr.Add(errors.WrapConfigError("trackpad", "add_command", "FirstClickThreshold", t.ClickWeight, err))
		}
	}

	if t.ThreeFingerDrag != nil {
		batch.AddBool(trackpadDomain, "TrackpadThreeFingerDrag", *t.ThreeFingerDrag)
	}

	log.Debug("Applying trackpad defaults")
	if err := batch.Execute(ctx, log); err != nil {
		return errors.WrapConfigError("trackpad", "execute_batch", "", nil, err)
	}

	log.Debug("Trackpad configuration applied successfully")
	return nil
}
