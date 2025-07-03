package config

import (
	"context"
	"os/exec"

	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
)

type Executor interface {
	Execute(ctx context.Context) error
}

type LoggingExecutor interface {
	Execute(ctx context.Context, logger *logger.Logger) error
}

type ConfigExecutor interface {
	Validator
	Stringer
	Executor
}

type CommandExecutor interface {
	Execute(ctx context.Context) error
}

type BatchExecutor interface {
	AddCommand(cmd defaults.Command)
	Execute(ctx context.Context, logger *logger.Logger) error
}

type ProcessRestarter interface {
	Execute(ctx context.Context) error
}

type DefaultsWriter interface {
	WriteBool(domain, key string, value bool) error
	WriteInt(domain, key string, value int64) error
	WriteFloat(domain, key string, value float64) error
	WriteString(domain, key string, value string) error
	WriteEnum(domain, key string, value string, validValues []string) error
}

type SystemExecutor interface {
	RunCommand(ctx context.Context, name string, args ...string) error
	RunCommandWithOutput(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ConfigManager interface {
	RegisterConfig(domain string, config any) error
	GetConfig(domain string) (any, error)
	ValidateConfig(domain string, config any) error
	ExecuteConfig(ctx context.Context, domain string, config any) error
}

type DependencyContainer interface {
	GetLogger() *logger.Logger
	GetCommandExecutor() CommandExecutor
	GetBatchExecutor() BatchExecutor
	GetProcessRestarter() ProcessRestarter
	GetSystemExecutor() SystemExecutor
}

type ConfigService struct {
	logger           *logger.Logger
	batchExecutor    BatchExecutor
	processRestarter ProcessRestarter
	systemExecutor   SystemExecutor
}

func NewConfigService(
	logger *logger.Logger,
	batchExecutor BatchExecutor,
	processRestarter ProcessRestarter,
	systemExecutor SystemExecutor,
) *ConfigService {
	return &ConfigService{
		logger:           logger,
		batchExecutor:    batchExecutor,
		processRestarter: processRestarter,
		systemExecutor:   systemExecutor,
	}
}

func (s *ConfigService) GetLogger() *logger.Logger {
	return s.logger
}

func (s *ConfigService) GetBatchExecutor() BatchExecutor {
	return s.batchExecutor
}

func (s *ConfigService) GetProcessRestarter() ProcessRestarter {
	return s.processRestarter
}

func (s *ConfigService) GetSystemExecutor() SystemExecutor {
	return s.systemExecutor
}

func (s *ConfigService) ExecuteWithLogging(ctx context.Context, executor LoggingExecutor) error {
	return executor.Execute(ctx, s.logger)
}

type ComponentFactory interface {
	CreateBatchExecutor() BatchExecutor
	CreateProcessRestarter() ProcessRestarter
	CreateSystemExecutor() SystemExecutor
}

type DefaultComponentFactory struct{}

func NewDefaultComponentFactory() *DefaultComponentFactory {
	return &DefaultComponentFactory{}
}

func (f *DefaultComponentFactory) CreateBatchExecutor() BatchExecutor {
	return &BatchExecutorWrapper{
		executor: defaults.NewBatchExecutor(),
	}
}

func (f *DefaultComponentFactory) CreateProcessRestarter() ProcessRestarter {
	return defaults.NewKillallExecutor("")
}

func (f *DefaultComponentFactory) CreateSystemExecutor() SystemExecutor {
	return &DefaultSystemExecutor{}
}

type BatchExecutorWrapper struct {
	executor *defaults.BatchExecutor
}

func (w *BatchExecutorWrapper) AddCommand(cmd defaults.Command) {
	w.executor.AddCommand(cmd)
}

func (w *BatchExecutorWrapper) Execute(ctx context.Context, logger *logger.Logger) error {
	return w.executor.Execute(ctx, logger)
}

type DefaultSystemExecutor struct{}

func (e *DefaultSystemExecutor) RunCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Run()
}

func (e *DefaultSystemExecutor) RunCommandWithOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

func NewConfigServiceWithDefaults(logger *logger.Logger) *ConfigService {
	factory := NewDefaultComponentFactory()

	return NewConfigService(
		logger,
		factory.CreateBatchExecutor(),
		factory.CreateProcessRestarter(),
		factory.CreateSystemExecutor(),
	)
}
