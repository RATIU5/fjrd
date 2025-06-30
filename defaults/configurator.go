package defaults

import "fmt"

// Configurator represents any system component that can be configured
type Configurator interface {
	Validate() error
	Execute() error
}

// ConfigurableGroup represents a group of configurators that can be executed together
type ConfigurableGroup struct {
	name          string
	configurators []Configurator
}

// NewConfigurableGroup creates a new configurable group
func NewConfigurableGroup(name string) *ConfigurableGroup {
	return &ConfigurableGroup{
		name:          name,
		configurators: make([]Configurator, 0),
	}
}

// Add adds a configurator to the group
func (g *ConfigurableGroup) Add(configurator Configurator) {
	g.configurators = append(g.configurators, configurator)
}

// Validate validates all configurators in the group
func (g *ConfigurableGroup) Validate() error {
	for i, configurator := range g.configurators {
		if err := configurator.Validate(); err != nil {
			return fmt.Errorf("validation failed for %s configurator %d: %w", g.name, i, err)
		}
	}
	return nil
}

// Execute executes all configurators in the group
func (g *ConfigurableGroup) Execute() error {
	for i, configurator := range g.configurators {
		if err := configurator.Execute(); err != nil {
			return fmt.Errorf("execution failed for %s configurator %d: %w", g.name, i, err)
		}
	}
	return nil
}

// MacOSConfiguration represents a complete macOS configuration
type MacOSConfiguration struct {
	groups []*ConfigurableGroup
}

// NewMacOSConfiguration creates a new macOS configuration
func NewMacOSConfiguration() *MacOSConfiguration {
	return &MacOSConfiguration{
		groups: make([]*ConfigurableGroup, 0),
	}
}

// AddGroup adds a configurable group to the configuration
func (c *MacOSConfiguration) AddGroup(group *ConfigurableGroup) {
	c.groups = append(c.groups, group)
}

// Validate validates all groups in the configuration
func (c *MacOSConfiguration) Validate() error {
	for _, group := range c.groups {
		if err := group.Validate(); err != nil {
			return fmt.Errorf("macOS configuration validation failed: %w", err)
		}
	}
	return nil
}

// Execute executes all groups in the configuration
func (c *MacOSConfiguration) Execute() error {
	for _, group := range c.groups {
		if err := group.Execute(); err != nil {
			return fmt.Errorf("macOS configuration execution failed: %w", err)
		}
	}
	return nil
}
