package desktop

import (
	"context"

	"github.com/RATIU5/fjrd/internal/errors"
	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/shared"
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
	return shared.FormatConfig("Desktop", d)
}

func (d *Config) Fields() map[string]any {
	fields := make(map[string]any)

	if d.SortFoldersFirst != nil {
		fields["sort-folders-first"] = *d.SortFoldersFirst
	}
	if d.ShowIcons != nil {
		fields["show-icons"] = *d.ShowIcons
	}
	if d.ShowHardDrives != nil {
		fields["show-hard-drives"] = *d.ShowHardDrives
	}
	if d.ShowExternalHardDrives != nil {
		fields["show-external-hard-drives"] = *d.ShowExternalHardDrives
	}
	if d.ShowRemovableMedia != nil {
		fields["show-removable-media"] = *d.ShowRemovableMedia
	}
	if d.ShowMountedServers != nil {
		fields["show-mounted-servers"] = *d.ShowMountedServers
	}

	return fields
}

func (d *Config) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("desktop")
	log.Debug("Configuring desktop settings")

	multiErr := &errors.MultiError{}
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
		multiErr.Add(errors.WrapConfigError("desktop", "batch_execute", "desktop_defaults", nil, err))
	}

	if err := multiErr.ToError(); err != nil {
		return err
	}

	log.Debug("Restarting Finder to apply changes")
	killall := defaults.NewKillallExecutor("Finder")
	if err := killall.Execute(ctx); err != nil {
		return errors.WrapConfigError("desktop", "restart", "finder_process", nil, err)
	}

	log.Debug("Desktop configuration applied successfully")
	return nil
}
