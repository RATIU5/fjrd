package finder

import (
	"context"
	"fmt"
	"strings"

	"github.com/RATIU5/fjrd/internal/macos/defaults"
)

type Config struct {
	ShowAllExtensions             *bool               `toml:"show-all-extensions,omitempty"`
	ShowAllFiles                  *bool               `toml:"show-all-files,omitempty"`
	ShowPathBar                   *bool               `toml:"show-path-bar,omitempty"`
	PreferredViewStyle            *PreferredViewStyle `toml:"preferred-view-style,omitmepty"`
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
	var parts []string
	if f.ShowAllExtensions != nil {
		parts = append(parts, fmt.Sprintf("show-all-extensions: %t", *f.ShowAllExtensions))
	}
	if f.ShowAllFiles != nil {
		parts = append(parts, fmt.Sprintf("show-all-files: %t", *f.ShowAllFiles))
	}
	if f.ShowPathBar != nil {
		parts = append(parts, fmt.Sprintf("show-path-bar: %t", *f.ShowPathBar))
	}
	if f.PreferredViewStyle != nil {
		parts = append(parts, fmt.Sprintf("preferred-view-style: %s", *f.PreferredViewStyle))
	}
	if f.SortFoldersFirst != nil {
		parts = append(parts, fmt.Sprintf("sort-folders-first: %t", *f.SortFoldersFirst))
	}
	if f.FinderSpawnTab != nil {
		parts = append(parts, fmt.Sprintf("finder-spawn-tab: %t", *f.FinderSpawnTab))
	}
	if f.DefaultSearchScope != nil {
		parts = append(parts, fmt.Sprintf("default-search-scope: %s", *f.DefaultSearchScope))
	}
	if f.RemoveOldTrashItems != nil {
		parts = append(parts, fmt.Sprintf("remove-old-trash-items: %t", *f.RemoveOldTrashItems))
	}
	if f.ShowExtensionChangeWarning != nil {
		parts = append(parts, fmt.Sprintf("show-extension-change-warning: %t", *f.ShowExtensionChangeWarning))
	}
	if f.SaveNewDocsToCloud != nil {
		parts = append(parts, fmt.Sprintf("save-new-docs-to-cloud: %t", *f.SaveNewDocsToCloud))
	}
	if f.ShowWindowTitlebarIcons != nil {
		parts = append(parts, fmt.Sprintf("show-window-titlebar-icons: %t", *f.ShowWindowTitlebarIcons))
	}
	if f.ToolbarTitleViewRolloverDelay != nil {
		parts = append(parts, fmt.Sprintf("toolbar-title-view-rollover-delay: %.2f", *f.ToolbarTitleViewRolloverDelay))
	}
	if f.TableViewDefaultSizeMode != nil {
		parts = append(parts, fmt.Sprintf("table-view-default-size-mode: %d", *f.TableViewDefaultSizeMode))
	}

	if len(parts) == 0 {
		return "Finder{}"
	}

	return fmt.Sprintf("Finder{%s}", strings.Join(parts, ", "))
}

func (f *Config) Execute(ctx context.Context, log interface {
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}) error {
	log.Debug("Configuring finder settings")
	batch := defaults.NewBatchExecutor()
	const finderDomain = "com.apple.finder"
	const nsGlobalDomain = "NSGlobalDomain"
	const universalDomain = "com.apple.universalaccess"

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
			return fmt.Errorf("failed to add toolbar-title-view-rollover-delay command: %w", err)
		}
	}

	if f.TableViewDefaultSizeMode != nil {
		if err := batch.AddInt(nsGlobalDomain, "NSTableViewDefaultSizeMode", *f.TableViewDefaultSizeMode); err != nil {
			return fmt.Errorf("failed to add table-view-default-size-mode command: %w", err)
		}
	}

	log.Debug("Applying finder defaults")
	if err := batch.Execute(ctx, log); err != nil {
		return fmt.Errorf("failed to execute finder configuration: %w", err)
	}

	log.Debug("Restarting Finder to apply changes")
	killall := defaults.NewKillallExecutor("Finder")
	if err := killall.Execute(ctx); err != nil {
		return fmt.Errorf("failed to restart finder: %w", err)
	}

	log.Debug("Finder configuration applied successfully")
	return nil
}
