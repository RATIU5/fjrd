package safari

import (
	"context"
	"fmt"
	"strings"

	"github.com/RATIU5/fjrd/internal/macos/defaults"
)

type Config struct {
	ShowFullUrl *bool `toml:"show-full-url,omitempty"`
}

func (s *Config) Validate() error {
	return nil
}

func (s *Config) String() string {
	var parts []string

	if s.ShowFullUrl != nil {
		parts = append(parts, fmt.Sprintf("show-full-url: %t", *s.ShowFullUrl))
	}

	if len(parts) == 0 {
		return "Safari{}"
	}

	return fmt.Sprintf("Safari{%s}", strings.Join(parts, ", "))
}

func (s *Config) Execute(ctx context.Context, log interface {
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}) error {
	log.Debug("Configuring safari settings")
	batch := defaults.NewBatchExecutor()
	const safariDomain = "com.apple.Safari"

	if s.ShowFullUrl != nil {
		batch.AddBool(safariDomain, "ShowFullURLInSmartSearchField", *s.ShowFullUrl)
	}

	log.Debug("Applying safari defaults")
	if err := batch.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute safari configuration: %w", err)
	}

	log.Debug("Restarting safari to apply changes")
	killall := defaults.NewKillallExecutor("Safari")
	if err := killall.ExecuteIfRunning(ctx); err != nil {
		return fmt.Errorf("failed to restart safari: %w", err)
	}

	log.Debug("Safari configuration applied successfully")
	return nil
}
