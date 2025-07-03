package finder

import (
	"fmt"
	"strings"
)

type PreferredViewStyle string
type DefaultSearchScope string

const (
	ColumnView  PreferredViewStyle = "clmv"
	ListView    PreferredViewStyle = "Nlsv"
	GallaryView PreferredViewStyle = "glyv"
	IconView    PreferredViewStyle = "icnv"
)

const (
	CurrentFolder  DefaultSearchScope = "SCcf"
	PreviousSearch DefaultSearchScope = "SCsp"
	SearchMac      DefaultSearchScope = "SCev"
)

var preferredViewStyleAliases = map[string]PreferredViewStyle{
	"column":  ColumnView,
	"list":    ListView,
	"gallery": GallaryView,
	"icon":    IconView,
	"clmv":    ColumnView,
	"nlsv":    ListView,
	"glyv":    GallaryView,
	"icnv":    IconView,
}

var defaultSearchScopeAliases = map[string]DefaultSearchScope{
	"current":  CurrentFolder,
	"previous": PreviousSearch,
	"mac":      SearchMac,
	"sccf":     CurrentFolder,
	"scsp":     PreviousSearch,
	"scev":     SearchMac,
}

func (p PreferredViewStyle) String() string {
	return string(p)
}

func (p PreferredViewStyle) IsValid() bool {
	switch p {
	case ColumnView, ListView, GallaryView, IconView:
		return true
	default:
		return false
	}
}

func (s DefaultSearchScope) String() string {
	return string(s)
}

func (s DefaultSearchScope) IsValid() bool {
	switch s {
	case CurrentFolder, PreviousSearch, SearchMac:
		return true
	default:
		return false
	}
}

func ParsePreferredViewStyle(s string) (PreferredViewStyle, error) {
	normalized := strings.ToLower(strings.TrimSpace(s))

	if view, exists := preferredViewStyleAliases[normalized]; exists {
		return view, nil
	}

	view := PreferredViewStyle(s)
	if view.IsValid() {
		return view, nil
	}

	var validOptions []string
	for alias := range preferredViewStyleAliases {
		if len(alias) > 4 {
			validOptions = append(validOptions, alias)
		}
	}

	return "", fmt.Errorf("invalid preferred-view-style %q, must be one of: %s", s, strings.Join(validOptions, ", "))
}

func ParseDefaultSearchScope(s string) (DefaultSearchScope, error) {
	normalized := strings.ToLower(strings.TrimSpace(s))

	if scope, exists := defaultSearchScopeAliases[normalized]; exists {
		return scope, nil
	}

	searchScope := DefaultSearchScope(s)
	if searchScope.IsValid() {
		return searchScope, nil
	}

	var validOptions []string
	for alias := range defaultSearchScopeAliases {
		if len(alias) > 4 {
			validOptions = append(validOptions, alias)
		}
	}

	return "", fmt.Errorf("invalid default-search-scope %q, must be one of: %s", s, strings.Join(validOptions, ", "))
}

func AllPreferredViewStyles() []PreferredViewStyle {
	return []PreferredViewStyle{ColumnView, ListView, GallaryView, IconView}
}

func AllDefaultSearchScopes() []DefaultSearchScope {
	return []DefaultSearchScope{CurrentFolder, PreviousSearch, SearchMac}
}

func PreferredViewStyleAliases() []string {
	var aliases []string
	for alias := range preferredViewStyleAliases {
		if len(alias) > 4 {
			aliases = append(aliases, alias)
		}
	}
	return aliases
}

func DefaultSearchScopeAliases() []string {
	var aliases []string
	for alias := range defaultSearchScopeAliases {
		if len(alias) > 4 {
			aliases = append(aliases, alias)
		}
	}
	return aliases
}

func (p *PreferredViewStyle) UnmarshalText(text []byte) error {
	parsed, err := ParsePreferredViewStyle(string(text))
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}

func (s *DefaultSearchScope) UnmarshalText(text []byte) error {
	parsed, err := ParseDefaultSearchScope(string(text))
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}
