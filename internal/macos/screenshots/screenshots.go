package screenshots

import (
	"context"
	"fmt"

	"github.com/RATIU5/fjrd/internal/errors"
	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/shared"
)

type Config struct {
	DisableShadow *bool   `toml:"disable-shadow,omitempty"`
	IncludeDate   *bool   `toml:"include-date,omitempty"`
	SaveLocation  *string `toml:"save-location,omitempty"`
	ShowThumbnail *bool   `toml:"show-thumbnail,omitempty"`
	Format        *Format `toml:"format,omitempty"`
}

func (s *Config) Validate() error {
	if s.Format != nil && !s.Format.IsValid() {
		return errors.WrapConfigError("screenshots", "validate", "format", *s.Format, fmt.Errorf("invalid format: %s", *s.Format))
	}
	return nil
}

func (s *Config) String() string {
	return shared.FormatConfig("Screenshots", s)
}

func (s *Config) Fields() map[string]any {
	fields := make(map[string]any)

	if s.DisableShadow != nil {
		fields["disable-shadow"] = *s.DisableShadow
	}
	if s.IncludeDate != nil {
		fields["include-date"] = *s.IncludeDate
	}
	if s.SaveLocation != nil {
		fields["save-location"] = *s.SaveLocation
	}
	if s.ShowThumbnail != nil {
		fields["show-thumbnail"] = *s.ShowThumbnail
	}
	if s.Format != nil {
		fields["format"] = s.Format.String()
	}

	return fields
}

func (s *Config) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("screenshots")
	log.Debug("Configuring screenshot settings")

	multiErr := &errors.MultiError{}
	batch := defaults.NewBatchExecutor()
	const screenshotDomain = "com.apple.screencapture"

	if s.DisableShadow != nil {
		batch.AddBool(screenshotDomain, "disable-shadow", *s.DisableShadow)
	}

	if s.IncludeDate != nil {
		batch.AddBool(screenshotDomain, "include-date", *s.IncludeDate)
	}

	if s.SaveLocation != nil {
		batch.AddString(screenshotDomain, "location", *s.SaveLocation)
	}

	if s.ShowThumbnail != nil {
		batch.AddBool(screenshotDomain, "show-thumbnail", *s.ShowThumbnail)
	}

	if s.Format != nil {
		format := defaults.NewEnumValue(string(s.Format.String()), []string{"png", "jpg", "jpeg", "pdf", "psd", "gif", "tga", "tiff", "bmp", "heic"})
		batch.AddCommand(defaults.Command{
			Domain: screenshotDomain,
			Key:    "type",
			Value:  format,
		})
	}

	log.Debug("Applying screenshot defaults")
	if err := batch.Execute(ctx, log); err != nil {
		multiErr.Add(errors.WrapConfigError("screenshots", "batch_execute", "screenshot_defaults", nil, err))
	}

	if err := multiErr.ToError(); err != nil {
		return err
	}

	log.Debug("Screenshot configuration applied successfully")
	return nil
}
