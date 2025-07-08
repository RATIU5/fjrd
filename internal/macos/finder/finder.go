package finder

import (
	"context"
	"fmt"

	"github.com/RATIU5/fjrd/internal/errors"
	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
	"github.com/RATIU5/fjrd/internal/shared"
)

type Config struct {
	ShowAllExtensions             *bool               `toml:"show-all-extensions,omitempty"`
	ShowAllFiles                  *bool               `toml:"show-all-files,omitempty"`
	ShowPathBar                   *bool               `toml:"show-path-bar,omitempty"`
	PreferredViewStyle            *PreferredViewStyle `toml:"preferred-view-style,omitempty"`
	SortFoldersFirst              *bool               `toml:"sort-folders-first,omitempty"`
	FinderSpawnTab                *bool               `toml:"finder-spawn-tab,omitempty"`
	DefaultSearchScope            *DefaultSearchScope `toml:"default-search-scope,omitempty"`
	RemoveOldTrashItems           *bool               `toml:"remove-old-trash-items,omitempty"`
	ShowExtensionChangeWarning    *bool               `toml:"show-extension-change-warning,omitempty"`
	SaveNewDocsToCloud            *bool               `toml:"save-new-docs-to-cloud,omitempty"`
	ShowWindowTitlebarIcons       *bool               `toml:"show-window-titlebar-icons,omitempty"`
	ToolbarTitleViewRolloverDelay *float32            `toml:"toolbar-title-view-rollover-delay,omitempty"`
	TableViewDefaultSizeMode      *int16              `toml:"table-view-default-size-mode,omitempty"`
}

func (f *Config) Validate() error {
	if f.PreferredViewStyle != nil && !f.PreferredViewStyle.IsValid() {
		return fmt.Errorf("invalid preferred-view-style: %s", *f.PreferredViewStyle)
	}
	if f.DefaultSearchScope != nil && !f.DefaultSearchScope.IsValid() {
		return fmt.Errorf("invalid default-search-scope value: %s", *f.DefaultSearchScope)
	}
	return nil
}

func (f *Config) String() string {
	return shared.FormatConfig("Finder", f)
}

func (f *Config) Fields() map[string]any {
	fields := make(map[string]any)

	if f.ShowAllExtensions != nil {
		fields["show-all-extensions"] = f.ShowAllExtensions
	}
	if f.ShowAllFiles != nil {
		fields["show-all-files"] = f.ShowAllFiles
	}
	if f.ShowPathBar != nil {
		fields["show-path-bar"] = f.ShowPathBar
	}
	if f.PreferredViewStyle != nil {
		fields["preferred-view-style"] = f.PreferredViewStyle
	}
	if f.SortFoldersFirst != nil {
		fields["sort-folders-first"] = f.SortFoldersFirst
	}
	if f.FinderSpawnTab != nil {
		fields["finder-spawn-tab"] = f.FinderSpawnTab
	}
	if f.DefaultSearchScope != nil {
		fields["default-search-scope"] = f.DefaultSearchScope
	}
	if f.RemoveOldTrashItems != nil {
		fields["remove-old-trash-items"] = f.RemoveOldTrashItems
	}
	if f.ShowExtensionChangeWarning != nil {
		fields["show-extension-change-warning"] = f.ShowExtensionChangeWarning
	}
	if f.SaveNewDocsToCloud != nil {
		fields["save-new-docs-to-cloud"] = f.SaveNewDocsToCloud
	}
	if f.ShowWindowTitlebarIcons != nil {
		fields["show-window-titlebar-icons"] = f.ShowWindowTitlebarIcons
	}
	if f.ToolbarTitleViewRolloverDelay != nil {
		fields["toolbar-title-view-rollover-delay"] = f.ToolbarTitleViewRolloverDelay
	}
	if f.TableViewDefaultSizeMode != nil {
		fields["table-view-default-size-mode"] = f.TableViewDefaultSizeMode
	}

	return fields
}

func (f *Config) Execute(ctx context.Context, log *logger.Logger) error {
	log = log.WithComponent("finder")
	log.Debug("Configuring finder settings")

	batch := defaults.NewBatchExecutor()
	const finderDomain = "com.apple.finder"
	const nsGlobalDomain = "NSGlobalDomain"
	const universalDomain = "com.apple.universalaccess"

	multiErr := errors.NewMultiError()

	if f.ShowAllExtensions != nil {
		batch.AddBool(nsGlobalDomain, "AppleShowAllExtensions", *f.ShowAllExtensions)
	}

	if f.ShowAllFiles != nil {
		batch.AddBool(finderDomain, "AppleShowAllFiles", *f.ShowAllFiles)
	}

	if f.ShowPathBar != nil {
		batch.AddBool(finderDomain, "ShowPathbar", *f.ShowPathBar)
	}

	if f.PreferredViewStyle != nil {
		value := defaults.NewEnumValue(string(*f.PreferredViewStyle), []string{"clmv", "Nlsv", "glyv", "icnv"})
		batch.AddCommand(defaults.Command{
			Domain: finderDomain,
			Key:    "FXPreferredViewStyle",
			Value:  value,
		})
	}

	if f.SortFoldersFirst != nil {
		batch.AddBool(finderDomain, "_FXSortFoldersFirst", *f.SortFoldersFirst)
	}

	if f.FinderSpawnTab != nil {
		batch.AddBool(finderDomain, "FinderSpawnTab", *f.FinderSpawnTab)
	}

	if f.DefaultSearchScope != nil {
		value := defaults.NewEnumValue(string(*f.DefaultSearchScope), []string{"SCcf", "SCsp", "SCev"})
		batch.AddCommand(defaults.Command{
			Domain: finderDomain,
			Key:    "FXDefaultSearchScope",
			Value:  value,
		})
	}

	if f.RemoveOldTrashItems != nil {
		batch.AddBool(finderDomain, "FXRemoveOldTrashItems", *f.RemoveOldTrashItems)
	}

	if f.ShowExtensionChangeWarning != nil {
		batch.AddBool(finderDomain, "FXEnableExtensionChangeWarning", *f.ShowExtensionChangeWarning)
	}

	if f.SaveNewDocsToCloud != nil {
		batch.AddBool(nsGlobalDomain, "NSDocumentSaveNewDocumentsToCloud", *f.SaveNewDocsToCloud)
	}

	if f.ShowWindowTitlebarIcons != nil {
		batch.AddBool(universalDomain, "showWindowTitlebarIcons", *f.ShowWindowTitlebarIcons)
	}

	if f.ToolbarTitleViewRolloverDelay != nil {
		if err := batch.AddFloat(nsGlobalDomain, "NSToolbarTitleViewRolloverDelay", *f.ToolbarTitleViewRolloverDelay); err != nil {
			multiErr.Add(errors.WrapConfigError("finder", "add_float", "NSToolbarTitleViewRolloverDelay", *f.ToolbarTitleViewRolloverDelay, err))
		}
	}

	if f.TableViewDefaultSizeMode != nil {
		if err := batch.AddInt(nsGlobalDomain, "NSTableViewDefaultSizeMode", *f.TableViewDefaultSizeMode); err != nil {
			multiErr.Add(errors.WrapConfigError("finder", "add_int", "NSTableViewDefaultSizeMode", *f.TableViewDefaultSizeMode, err))
		}
	}

	if err := multiErr.ToError(); err != nil {
		return err
	}

	log.Debug("Applying finder defaults")
	if err := batch.Execute(ctx, log); err != nil {
		return errors.WrapConfigError("finder", "execute_batch", "", nil, err)
	}

	log.Debug("Restarting Finder to apply changes")
	killall := defaults.NewKillallExecutor("Finder")
	if err := killall.Execute(ctx); err != nil {
		return errors.WrapConfigError("finder", "restart_process", "Dock", nil, err)
	}

	log.Debug("Finder configuration applied successfully")
	return nil
}
