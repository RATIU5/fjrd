package safari

import (
	"context"

	"github.com/RATIU5/fjrd/internal/errors"
	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/shared"
)

type Config struct {
	ShowFullUrl *bool `toml:"show-full-url,omitempty"`
}

func (s *Config) Validate() error {
	return nil
}

func (s *Config) String() string {
	return shared.FormatConfig("Safari", s)
}

func (s *Config) Fields() map[string]any {
	fields := make(map[string]any)

	if s.ShowFullUrl != nil {
		fields["show-full-url"] = *s.ShowFullUrl
	}

	return fields
}

func (s *Config) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("safari")
	log.Debug("Configuring safari settings")

	multiErr := &errors.MultiError{}
	batch := defaults.NewBatchExecutor()
	const safariDomain = "com.apple.Safari"

	if s.ShowFullUrl != nil {
		batch.AddBool(safariDomain, "ShowFullURLInSmartSearchField", *s.ShowFullUrl)
	}

	log.Debug("Applying safari defaults")
	if err := batch.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("safari", "batch_execute", "safari_defaults", nil, err))
	}

	if err := multiErr.ToError(); err != nil {
		return err
	}

	log.Debug("Restarting safari to apply changes")
	killall := defaults.NewKillallExecutor("Safari")
	if err := killall.ExecuteIfRunning(ctx); err != nil {
		return errors.WrapConfigError("safari", "restart", "safari_process", nil, err)
	}

	log.Debug("Safari configuration applied successfully")
	return nil
}
