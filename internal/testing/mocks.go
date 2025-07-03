package testing

import (
	"context"
	"sync"

	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/defaults"
)

type MockLogger struct {
	mu      sync.RWMutex
	entries []LogEntry
}

type LogEntry struct {
	Level     string
	Message   string
	Fields    map[string]any
	Component string
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		entries: make([]LogEntry, 0),
	}
}

func (m *MockLogger) Info(msg string, args ...any) {
	m.log("INFO", msg, parseArgs(args...))
}

func (m *MockLogger) Debug(msg string, args ...any) {
	m.log("DEBUG", msg, parseArgs(args...))
}

func (m *MockLogger) Warn(msg string, args ...any) {
	m.log("WARN", msg, parseArgs(args...))
}

func (m *MockLogger) Error(msg string, args ...any) {
	m.log("ERROR", msg, parseArgs(args...))
}

func (m *MockLogger) With(args ...any) *logger.Logger {
	return &logger.Logger{}
}

func (m *MockLogger) WithComponent(component string) *logger.Logger {
	return &logger.Logger{}
}

func (m *MockLogger) WithOperation(operation string) *logger.Logger {
	return &logger.Logger{}
}

func (m *MockLogger) WithError(err error) *logger.Logger {
	return &logger.Logger{}
}

func (m *MockLogger) log(level, msg string, fields map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.entries = append(m.entries, LogEntry{
		Level:   level,
		Message: msg,
		Fields:  fields,
	})
}

func (m *MockLogger) GetEntries() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	entries := make([]LogEntry, len(m.entries))
	copy(entries, m.entries)
	return entries
}

func (m *MockLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = m.entries[:0]
}

func parseArgs(args ...any) map[string]any {
	fields := make(map[string]any)
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			fields[key] = args[i+1]
		}
	}
	return fields
}

type MockBatchExecutor struct {
	mu       sync.RWMutex
	commands []defaults.Command
	executed bool
	execErr  error
}

func NewMockBatchExecutor() *MockBatchExecutor {
	return &MockBatchExecutor{
		commands: make([]defaults.Command, 0),
	}
}

func (m *MockBatchExecutor) AddCommand(cmd defaults.Command) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commands = append(m.commands, cmd)
}

func (m *MockBatchExecutor) AddBool(domain, key string, value bool) {
	m.AddCommand(defaults.Command{
		Domain: domain,
		Key:    key,
		Value:  defaults.NewBoolValue(value),
	})
}

func (m *MockBatchExecutor) AddInt(domain, key string, value int16) error {
	intValue, err := defaults.NewIntValue(int64(value))
	if err != nil {
		return err
	}
	m.AddCommand(defaults.Command{
		Domain: domain,
		Key:    key,
		Value:  intValue,
	})
	return nil
}

func (m *MockBatchExecutor) AddFloat(domain, key string, value float32) error {
	floatValue, err := defaults.NewFloatValue(float64(value))
	if err != nil {
		return err
	}
	m.AddCommand(defaults.Command{
		Domain: domain,
		Key:    key,
		Value:  floatValue,
	})
	return nil
}

func (m *MockBatchExecutor) Execute(ctx context.Context, log *logger.Logger) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.executed = true
	return m.execErr
}

func (m *MockBatchExecutor) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commands = m.commands[:0]
	m.executed = false
}

func (m *MockBatchExecutor) GetCommands() []defaults.Command {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	commands := make([]defaults.Command, len(m.commands))
	copy(commands, m.commands)
	return commands
}

func (m *MockBatchExecutor) WasExecuted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.executed
}

func (m *MockBatchExecutor) SetExecuteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.execErr = err
}

type MockProcessRestarter struct {
	mu        sync.RWMutex
	restarted []string
	execErr   error
}

func NewMockProcessRestarter() *MockProcessRestarter {
	return &MockProcessRestarter{
		restarted: make([]string, 0),
	}
}

func (m *MockProcessRestarter) Restart(ctx context.Context, processName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.restarted = append(m.restarted, processName)
	return m.execErr
}

func (m *MockProcessRestarter) Execute(ctx context.Context) error {
	return m.Restart(ctx, "")
}

func (m *MockProcessRestarter) GetRestartedProcesses() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	processes := make([]string, len(m.restarted))
	copy(processes, m.restarted)
	return processes
}

func (m *MockProcessRestarter) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.execErr = err
}

func (m *MockProcessRestarter) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.restarted = m.restarted[:0]
}

type MockSystemExecutor struct {
	mu       sync.RWMutex
	commands []SystemCommand
	execErr  error
}

type SystemCommand struct {
	Name string
	Args []string
}

func NewMockSystemExecutor() *MockSystemExecutor {
	return &MockSystemExecutor{
		commands: make([]SystemCommand, 0),
	}
}

func (m *MockSystemExecutor) RunCommand(ctx context.Context, name string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.commands = append(m.commands, SystemCommand{
		Name: name,
		Args: args,
	})
	
	return m.execErr
}

func (m *MockSystemExecutor) RunCommandWithOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	err := m.RunCommand(ctx, name, args...)
	return []byte("mock output"), err
}

func (m *MockSystemExecutor) GetCommands() []SystemCommand {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	commands := make([]SystemCommand, len(m.commands))
	copy(commands, m.commands)
	return commands
}

func (m *MockSystemExecutor) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.execErr = err
}

func (m *MockSystemExecutor) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commands = m.commands[:0]
}

type MockComponentFactory struct {
	BatchExecutor    *MockBatchExecutor
	ProcessRestarter *MockProcessRestarter
	SystemExecutor   *MockSystemExecutor
}

func NewMockComponentFactory() *MockComponentFactory {
	return &MockComponentFactory{
		BatchExecutor:    NewMockBatchExecutor(),
		ProcessRestarter: NewMockProcessRestarter(),
		SystemExecutor:   NewMockSystemExecutor(),
	}
}

type TestEnvironment struct {
	Logger           *MockLogger
	BatchExecutor    *MockBatchExecutor
	ProcessRestarter *MockProcessRestarter
	SystemExecutor   *MockSystemExecutor
}

func NewTestEnvironment() *TestEnvironment {
	mockLogger := NewMockLogger()
	mockBatch := NewMockBatchExecutor()
	mockRestart := NewMockProcessRestarter()
	mockSystem := NewMockSystemExecutor()
	
	return &TestEnvironment{
		Logger:           mockLogger,
		BatchExecutor:    mockBatch,
		ProcessRestarter: mockRestart,
		SystemExecutor:   mockSystem,
	}
}

func (te *TestEnvironment) Reset() {
	te.Logger.Clear()
	te.BatchExecutor.Clear()
	te.ProcessRestarter.Clear()
	te.SystemExecutor.Clear()
}