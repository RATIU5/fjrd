package keyboard

import (
	"context"
	"fmt"

	"github.com/RATIU5/fjrd/internal/errors"
	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/shared"
)

type Config struct {
	KeyHoldShowsAccents *bool       `toml:"key-hold-shows-accents,omitempty"`
	FnKeyBehavior       *FnBehavior `toml:"fn-key-behavior"`
	SpecialFKeys        *bool       `toml:"special-f-keys,omitempty"`
	TabNavigation       *bool       `toml:"tab-navigation,omitempty"`
	LanguageIndicator   *bool       `toml:"language-indicator,omitempty"`
}

func (k *Config) Validate() error {
	if k.FnKeyBehavior != nil && !k.FnKeyBehavior.IsValid() {
		return fmt.Errorf("invalid fn-key-behavior: %s", *k.FnKeyBehavior)
	}
	return nil
}

func (k *Config) String() string {
	return shared.FormatConfig("Keyboard", k)
}

func (k *Config) Fields() map[string]any {
	fields := make(map[string]any)

	if k.KeyHoldShowsAccents != nil {
		fields["key-hold-shows-accents"] = k.KeyHoldShowsAccents
	}
	if k.FnKeyBehavior != nil {
		fields["fn-key-behavior"] = k.FnKeyBehavior
	}
	if k.SpecialFKeys != nil {
		fields["special-f-keys"] = k.SpecialFKeys
	}
	if k.TabNavigation != nil {
		fields["tab-navigation"] = k.TabNavigation
	}
	if k.LanguageIndicator != nil {
		fields["language-indicator"] = k.LanguageIndicator
	}

	return fields
}

func (k *Config) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("keyboard")
	log.Debug("Configuring keyboard settings")

	batch := defaults.NewBatchExecutor()
	const globalDomain = "NSGlobalDomain"
	const toolboxDomain = "com.apple.HIToolbox"
	const prefDomain = "kCFPreferencesAnyApplication"

	multiErr := errors.NewMultiError()

	if k.KeyHoldShowsAccents != nil {
		batch.AddBool(globalDomain, "ApplePressAndHoldEnabled", *k.KeyHoldShowsAccents)
	}

	if k.FnKeyBehavior != nil {
		fnBehaviorValue := defaults.NewEnumValue(k.FnKeyBehavior.String(), []string{"dictation", "input-source", "emoji", "none"})
		batch.AddCommand(defaults.Command{
			Domain: toolboxDomain,
			Key:    "AppleFnUsageType",
			Value:  fnBehaviorValue,
		})
	}

	if k.SpecialFKeys != nil {
		batch.AddBool(globalDomain, "com.apple.keyboard.fnState", *k.SpecialFKeys)
	}

	if k.TabNavigation != nil {
		tabNavValue := NewTabNavigationValue(*k.TabNavigation)
		if err := batch.AddInt(globalDomain, "AppleKeyboardUIMode", tabNavValue.Convert().(int)); err != nil {
			multiErr.Add(errors.WrapConfigError("keyboard", "add_command", "AppleKeyboardUIMode", tabNavValue.Convert(), err))
		}
	}

	if k.LanguageIndicator != nil {
		batch.AddBool(prefDomain, "TSMLanguageIndicatorEnabled", *k.LanguageIndicator)
	}

	if err := multiErr.ToError(); err != nil {
		return err
	}

	log.Debug("Applying keyboard defaults")
	if err := batch.Execute(ctx, log); err != nil {
		return errors.WrapConfigError("keyboard", "execute_batch", "", nil, err)
	}

	log.Debug("Keyboard configuration applied successfully")
	return nil
}
