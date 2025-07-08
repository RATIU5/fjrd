package missionControl

import (
	"context"

	"github.com/RATIU5/fjrd/internal/errors"
	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/shared"
)

type Config struct {
	AutoRearrangeSpaces        *bool `toml:"auto-rearrange-spaces,omitempty"`
	GroupWindowsByApp          *bool `toml:"group-windows-by-app,omitempty"`
	SwitchToAppsOpenWindow     *bool `toml:"switch-to-apps-open-window,omitempty"`
	DisplaysHaveSeparateSpaces *bool `toml:"displays-have-separate-spaces,omitempty"`
}

func (m *Config) Validate() error {
	return nil
}

func (m *Config) String() string {
	return shared.FormatConfig("MissionControl", m)
}

func (m *Config) Fields() map[string]any {
	fields := make(map[string]any)

	if m.AutoRearrangeSpaces != nil {
		fields["auto-rearrange-spaces"] = m.AutoRearrangeSpaces
	}
	if m.GroupWindowsByApp != nil {
		fields["group-windows-by-app"] = m.GroupWindowsByApp
	}
	if m.SwitchToAppsOpenWindow != nil {
		fields["switch-to-apps-open-window"] = m.SwitchToAppsOpenWindow
	}
	if m.DisplaysHaveSeparateSpaces != nil {
		fields["displays-have-separate-spaces"] = m.DisplaysHaveSeparateSpaces
	}

	return fields
}

func (m *Config) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("mission-control")
	log.Debug("Configuring mission control settings")

	batch := defaults.NewBatchExecutor()
	const dockDomain = "com.apple.dock"
	const globalDomain = "NSGlobalDomain"
	const spacesDomain = "com.apple.spaces"

	if m.AutoRearrangeSpaces != nil {
		batch.AddBool(dockDomain, "mru-spaces", *m.AutoRearrangeSpaces)
	}

	if m.GroupWindowsByApp != nil {
		batch.AddBool(dockDomain, "expose-group-apps", *m.GroupWindowsByApp)
	}

	if m.SwitchToAppsOpenWindow != nil {
		batch.AddBool(globalDomain, "AppleSpacesSwitchOnActivate", *m.SwitchToAppsOpenWindow)
	}

	if m.DisplaysHaveSeparateSpaces != nil {
		batch.AddBool(spacesDomain, "spans-displays", *m.DisplaysHaveSeparateSpaces)
	}

	log.Debug("Applying mission control defaults")
	if err := batch.Execute(ctx, log); err != nil {
		return errors.WrapConfigError("mission-control", "execute_batch", "", nil, err)
	}

	log.Debug("Restarting dock to apply changes")
	killallDock := defaults.NewKillallExecutor("Dock")
	if err := killallDock.Execute(ctx); err != nil {
		return errors.WrapConfigError("dock", "restart_process", "Dock", nil, err)
	}

	log.Debug("Restarting system ui server to apply changes")
	killallUi := defaults.NewKillallExecutor("SystemUIServer")
	if err := killallUi.Execute(ctx); err != nil {
		return errors.WrapConfigError("system-ui-server", "restart_process", "SystemUIServer", nil, err)
	}

	log.Debug("Mission control configuration applied successfully")
	return nil
}
