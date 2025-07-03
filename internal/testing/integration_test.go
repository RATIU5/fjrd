package testing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/RATIU5/fjrd/internal/logger"
	"github.com/RATIU5/fjrd/internal/macos/dock"
)

func TestIntegrationWithRollback(t *testing.T) {
	if os.Getenv("FJRD_INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration tests. Set FJRD_INTEGRATION_TESTS=1 to run.")
	}

	log := logger.New(logger.LevelDebug, os.Stdout)
	testRunner := NewTestRunner(log)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	domains := []string{"com.apple.dock"}

	err := testRunner.RunWithRollback(ctx, domains, func() error {
		return testDockConfiguration(ctx, log)
	})

	if err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}
}

func testDockConfiguration(ctx context.Context, log *logger.Logger) error {
	config := &dock.Config{
		Autohide:    boolPtr(true),
		Orientation: positionPtr(dock.PositionLeft),
		TileSize:    int16Ptr(48),
	}

	if err := config.Validate(); err != nil {
		return err
	}

	return config.Execute(ctx, log)
}

func TestBackupAndRestore(t *testing.T) {
	if os.Getenv("FJRD_INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration tests. Set FJRD_INTEGRATION_TESTS=1 to run.")
	}

	log := logger.New(logger.LevelDebug, os.Stdout)
	backupManager := NewBackupManager(log, "")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	domains := []string{"com.apple.dock"}

	backup, err := backupManager.CreateBackup(ctx, domains)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	if backup.BackupID == "" {
		t.Error("Backup should have a valid ID")
	}

	if backup.DomainData == nil {
		t.Error("Backup should have domain data")
	}

	err = backupManager.RestoreBackup(ctx, backup)
	if err != nil {
		t.Fatalf("Failed to restore backup: %v", err)
	}
}

func TestMockEnvironment(t *testing.T) {
	testEnv := NewTestEnvironment()
	defer testEnv.Reset()

	if testEnv.Logger == nil {
		t.Error("Test environment should have a logger")
	}

	if testEnv.BatchExecutor == nil {
		t.Error("Test environment should have a batch executor")
	}

	if testEnv.ProcessRestarter == nil {
		t.Error("Test environment should have a process restarter")
	}

	if testEnv.SystemExecutor == nil {
		t.Error("Test environment should have a system executor")
	}

	testEnv.Logger.Info("test message", "key", "value")
	entries := testEnv.Logger.GetEntries()

	if len(entries) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(entries))
	}

	if entries[0].Level != "INFO" {
		t.Errorf("Expected INFO level, got %s", entries[0].Level)
	}

	if entries[0].Message != "test message" {
		t.Errorf("Expected 'test message', got %s", entries[0].Message)
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func int16Ptr(i int16) *int16 {
	return &i
}

func positionPtr(p dock.Position) *dock.Position {
	return &p
}
