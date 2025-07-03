package desktop

import (
	"context"
	"fmt"
	"strings"

	"github.com/RATIU5/fjrd/internal/macos/defaults"
)

type Config struct {
	SortFoldersFirst       *bool `toml:"sort-folders-first,omitempty"`
	ShowIcons              *bool `toml:"show-icons,omitempty"`
	ShowHardDrives         *bool `toml:"show-hard-drives,omitempty"`
	ShowExternalHardDrives *bool `toml:"show-external-hard-drives,omitempty"`
	ShowRemovableMedia     *bool `toml:"show-removable-media,omitempty"`
	ShowMountedServers     *bool `toml:"show-mounted-servers,omitempty"`
}

func (d *Config) Validate() error {
	return nil
}

func (d *Config) String() string {
	var parts []string
	if d.SortFoldersFirst != nil {
		parts = append(parts, fmt.Sprintf("sort-folders-first: %t", *d.SortFoldersFirst))
	}
	if d.ShowIcons != nil {
		parts = append(parts, fmt.Sprintf("show-icons: %t", *d.ShowIcons))
	}
	if d.ShowHardDrives != nil {
		parts = append(parts, fmt.Sprintf("show-hard-drives: %t", *d.ShowHardDrives))
	}
	if d.ShowExternalHardDrives != nil {
		parts = append(parts, fmt.Sprintf("show-external-hard-drives, %t", *d.ShowExternalHardDrives))
	}
	if d.ShowRemovableMedia != nil {
		parts = append(parts, fmt.Sprintf("show-removable-media: %t", *d.ShowRemovableMedia))
	}
	if d.ShowMountedServers != nil {
		parts = append(parts, fmt.Sprintf("show-mounted-servers: %t", *d.ShowMountedServers))
	}

	if len(parts) == 0 {
		return "Desktop{}"
	}

	return fmt.Sprintf("Desktop{%s}", strings.Join(parts, ", "))
}

func (d *Config) Execute(ctx context.Context, log interface {
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}) error {
	log.Debug("Configuring desktop settings")
	batch := defaults.NewBatchExecutor()
	const finderDomain = "com.apple.finder"

	if d.SortFoldersFirst != nil {
		batch.AddBool(finderDomain, "_FXSortFoldersFirstOnDesktop", *d.SortFoldersFirst)
	}
	if d.ShowIcons != nil {
		batch.AddBool(finderDomain, "CreateDesktop", *d.ShowIcons)
	}
	if d.ShowHardDrives != nil {
		batch.AddBool(finderDomain, "ShowHardDrivesOnDesktop", *d.ShowHardDrives)
	}
	if d.ShowExternalHardDrives != nil {
		batch.AddBool(finderDomain, "ShowExternalHardDrivesOnDesktop", *d.ShowExternalHardDrives)
	}
	if d.ShowRemovableMedia != nil {
		batch.AddBool(finderDomain, "ShowRemovableMediaOnDesktop", *d.ShowRemovableMedia)
	}
	if d.ShowMountedServers != nil {
		batch.AddBool(finderDomain, "ShowMountedServersOnDesktop", *d.ShowMountedServers)
	}

	log.Debug("Applying desktop defaults")
	if err := batch.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute desktop configuration: %w", err)
	}

	log.Debug("Restarting Finder to apply changes")
	killall := defaults.NewKillallExecutor("Finder")
	if err := killall.Execute(ctx); err != nil {
		return fmt.Errorf("failed to restart finder: %w", err)
	}

	log.Debug("Desktop configuration applied successfully")
	return nil
}
