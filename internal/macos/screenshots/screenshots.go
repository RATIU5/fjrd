package screenshots

import (
	"context"
	"fmt"
	"strings"

	"github.com/RATIU5/fjrd/internal/macos/defaults"
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
		return fmt.Errorf("invalid format: %s", *s.Format)
	}
	return nil
}

func (s *Config) String() string {
	var parts []string

	if s.DisableShadow != nil {
		parts = append(parts, fmt.Sprintf("disable-shadow: %t", *s.DisableShadow))
	}
	if s.IncludeDate != nil {
		parts = append(parts, fmt.Sprintf("include-date: %t", *s.IncludeDate))
	}
	if s.SaveLocation != nil {
		parts = append(parts, fmt.Sprintf("save-location: %s", *s.SaveLocation))
	}
	if s.ShowThumbnail != nil {
		parts = append(parts, fmt.Sprintf("show-thumbnail: %t", *s.ShowThumbnail))
	}
	if s.Format != nil {
		parts = append(parts, fmt.Sprintf("format: %s", s.Format.String()))
	}

	if len(parts) == 0 {
		return "Screenshot{}"
	}

	return fmt.Sprintf("Screenshot{%s}", strings.Join(parts, " ,"))
}

func (s *Config) Execute(ctx context.Context, log interface {
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}) error {
	log.Debug("Configuring screenshot settings")
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
		return fmt.Errorf("failed to execute screenshot configuration: %w", err)
	}

	log.Debug("Screenshot configuration applied successfully")
	return nil
}
