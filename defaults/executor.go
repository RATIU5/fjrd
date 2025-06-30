package defaults

import (
	"fmt"
	"os/exec"
)

// Executor handles macOS defaults command execution
type Executor interface {
	Execute() error
}

// CommandExecutor handles the execution of defaults commands
type CommandExecutor struct {
	domain string
}

// NewCommandExecutor creates a new executor for a specific domain
func NewCommandExecutor(domain string) *CommandExecutor {
	return &CommandExecutor{domain: domain}
}

// ValueType represents the type of value being set
type ValueType string

const (
	BoolType   ValueType = "-bool"
	StringType ValueType = "-string"
	IntType    ValueType = "-int"
	FloatType  ValueType = "-float"
)

// Command represents a defaults write command
type Command struct {
	Domain string
	Key    string
	Value  Value
}

// Execute runs the defaults command
func (c *Command) Execute() error {
	if err := c.Value.Validate(); err != nil {
		return NewValidationError(c.Key, c.Value, "", err)
	}

	args := []string{"write", c.Domain, c.Key, string(c.Value.Type()), c.Value.String()}
	cmd := exec.Command("defaults", args...)
	if err := cmd.Run(); err != nil {
		return NewExecutionError("defaults", args, err)
	}

	return nil
}


// BatchExecutor executes multiple commands in sequence
type BatchExecutor struct {
	commands []Command
}

// NewBatchExecutor creates a new batch executor
func NewBatchExecutor() *BatchExecutor {
	return &BatchExecutor{commands: make([]Command, 0)}
}

// AddCommand adds a command to the batch
func (b *BatchExecutor) AddCommand(cmd Command) {
	b.commands = append(b.commands, cmd)
}

// AddBool adds a boolean command to the batch
func (b *BatchExecutor) AddBool(domain, key string, value bool) {
	b.AddCommand(Command{
		Domain: domain,
		Key:    key,
		Value:  NewBoolValue(value),
	})
}

// AddString adds a string command to the batch
func (b *BatchExecutor) AddString(domain, key string, value string) {
	b.AddCommand(Command{
		Domain: domain,
		Key:    key,
		Value:  NewStringValue(value),
	})
}

// AddInt adds an integer command to the batch
func (b *BatchExecutor) AddInt(domain, key string, value interface{}) error {
	intValue, err := NewIntValue(value)
	if err != nil {
		return fmt.Errorf("failed to create int value for %s.%s: %w", domain, key, err)
	}
	b.AddCommand(Command{
		Domain: domain,
		Key:    key,
		Value:  intValue,
	})
	return nil
}

// AddFloat adds a float command to the batch
func (b *BatchExecutor) AddFloat(domain, key string, value interface{}) error {
	floatValue, err := NewFloatValue(value)
	if err != nil {
		return fmt.Errorf("failed to create float value for %s.%s: %w", domain, key, err)
	}
	b.AddCommand(Command{
		Domain: domain,
		Key:    key,
		Value:  floatValue,
	})
	return nil
}

// Execute runs all commands in the batch
func (b *BatchExecutor) Execute() error {
	for _, cmd := range b.commands {
		if err := cmd.Execute(); err != nil {
			return fmt.Errorf("batch execution failed: %w", err)
		}
	}
	return nil
}

// KillallExecutor executes a killall command to restart applications
type KillallExecutor struct {
	processName string
}

// NewKillallExecutor creates a new killall executor
func NewKillallExecutor(processName string) *KillallExecutor {
	return &KillallExecutor{processName: processName}
}

// Execute runs the killall command
func (k *KillallExecutor) Execute() error {
	args := []string{k.processName}
	cmd := exec.Command("killall", args...)
	if err := cmd.Run(); err != nil {
		return NewExecutionError("killall", args, err)
	}
	return nil
}