package defaults

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/RATIU5/fjrd/internal/errors"
)

type Executor interface {
	Execute(ctx context.Context) error
}

type CommandExecutor struct {
	domain string
}

func NewCommandExecutor(domain string) *CommandExecutor {
	return &CommandExecutor{domain: domain}
}

type Command struct {
	Domain string
	Key    string
	Value  Value
}

func (c *Command) Execute(ctx context.Context, log interface {
	Info(string, ...any)
	Debug(string, ...any)
}) error {
	if err := c.Value.Validate(); err != nil {
		return errors.NewValidationError(c.Key, c.Value, "", err)
	}

	if resetter, ok := c.Value.(ResetValue); ok && resetter.IsReset() {
		return c.executeReset(ctx, log)
	}

	args := []string{"write", c.Domain, c.Key, string(c.Value.Type()), c.Value.String()}
	cmd := exec.CommandContext(ctx, "defaults", args...)

	log.Debug("Executing defaults command",
		"domain", c.Domain,
		"key", c.Key,
		"type", string(c.Value.Type()),
		"value", c.Value.String(),
		"command", fmt.Sprintf("defaults %s", strings.Join(args, " ")))

	if err := cmd.Run(); err != nil {
		return errors.NewExecutionError("defaults", args, err)
	}

	return nil
}

func (c *Command) executeReset(ctx context.Context, log interface {
	Info(string, ...any)
	Debug(string, ...any)
}) error {
	args := []string{"delete", c.Domain, c.Key}
	cmd := exec.CommandContext(ctx, "defaults", args...)

	log.Debug("Resetting default to system value",
		"domain", c.Domain,
		"key", c.Key,
		"command", fmt.Sprintf("defaults %s", strings.Join(args, " ")))

	if err := cmd.Run(); err != nil {
		log.Debug("Reset failed (key may not exist)", "error", err)
	}

	return nil
}

type BatchExecutor struct {
	commands []Command
}

func NewBatchExecutor() *BatchExecutor {
	return &BatchExecutor{commands: make([]Command, 0)}
}

func (b *BatchExecutor) AddCommand(cmd Command) {
	b.commands = append(b.commands, cmd)
}

func (b *BatchExecutor) AddBool(domain, key string, value bool) {
	b.AddCommand(Command{
		Domain: domain,
		Key:    key,
		Value:  NewBoolValue(value),
	})
}

func (b *BatchExecutor) AddString(domain, key string, value string) {
	b.AddCommand(Command{
		Domain: domain,
		Key:    key,
		Value:  NewStringValue(value),
	})
}

func (b *BatchExecutor) AddInt(domain, key string, value any) error {
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

func (b *BatchExecutor) AddFloat(domain, key string, value any) error {
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

func (b *BatchExecutor) Execute(ctx context.Context, log interface {
	Info(string, ...any)
	Debug(string, ...any)
}) error {
	for _, cmd := range b.commands {
		if err := cmd.Execute(ctx, log); err != nil {
			return fmt.Errorf("batch execution failed: %w", err)
		}
	}
	return nil
}

type KillallExecutor struct {
	processName string
}

func NewKillallExecutor(processName string) *KillallExecutor {
	return &KillallExecutor{processName: processName}
}

func (k *KillallExecutor) Execute(ctx context.Context) error {
	args := []string{k.processName}
	cmd := exec.CommandContext(ctx, "killall", args...)
	if err := cmd.Run(); err != nil {
		return errors.NewExecutionError("killall", args, err)
	}
	return nil
}

func (k *KillallExecutor) ExecuteIfRunning(ctx context.Context) error {
	if !k.isProcessRunning(ctx) {
		return nil
	}
	return k.Execute(ctx)
}

func (k *KillallExecutor) isProcessRunning(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "pgrep", "-x", k.processName)
	err := cmd.Run()
	return err == nil
}
