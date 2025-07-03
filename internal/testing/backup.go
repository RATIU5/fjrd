package testing

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/RATIU5/fjrd/internal/logger"
)

type SystemBackup struct {
	BackupID    string
	Timestamp   time.Time
	BackupPath  string
	DomainData  map[string]map[string]any
	ProcessList []string
}

type BackupManager struct {
	logger     *logger.Logger
	backupDir  string
	maxBackups int
}

func NewBackupManager(logger *logger.Logger, backupDir string) *BackupManager {
	if backupDir == "" {
		homeDir, _ := os.UserHomeDir()
		backupDir = filepath.Join(homeDir, ".fjrd", "test-backups")
	}

	return &BackupManager{
		logger:     logger.WithComponent("backup"),
		backupDir:  backupDir,
		maxBackups: 10,
	}
}

func (bm *BackupManager) CreateBackup(ctx context.Context, domains []string) (*SystemBackup, error) {
	backupID := fmt.Sprintf("backup_%d", time.Now().Unix())
	timestamp := time.Now()

	backup := &SystemBackup{
		BackupID:   backupID,
		Timestamp:  timestamp,
		BackupPath: filepath.Join(bm.backupDir, backupID),
		DomainData: make(map[string]map[string]any),
	}

	bm.logger.Info("Creating system backup", "backup_id", backupID, "domains", len(domains))

	if err := os.MkdirAll(backup.BackupPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	for _, domain := range domains {
		domainData, err := bm.readDomainDefaults(ctx, domain)
		if err != nil {
			bm.logger.Warn("Failed to backup domain", "domain", domain, "error", err)
			continue
		}
		backup.DomainData[domain] = domainData
	}

	if err := bm.saveBackupMetadata(backup); err != nil {
		return nil, fmt.Errorf("failed to save backup metadata: %w", err)
	}

	bm.logger.Info("System backup created successfully", "backup_id", backupID)
	return backup, nil
}

func (bm *BackupManager) RestoreBackup(ctx context.Context, backup *SystemBackup) error {
	bm.logger.Info("Restoring system backup", "backup_id", backup.BackupID)

	for domain, data := range backup.DomainData {
		if err := bm.restoreDomainDefaults(ctx, domain, data); err != nil {
			bm.logger.Error("Failed to restore domain", "domain", domain, "error", err)
			return fmt.Errorf("failed to restore domain %s: %w", domain, err)
		}
	}

	bm.logger.Info("System backup restored successfully", "backup_id", backup.BackupID)
	return nil
}

func (bm *BackupManager) readDomainDefaults(ctx context.Context, domain string) (map[string]any, error) {
	cmd := exec.CommandContext(ctx, "defaults", "export", domain, "-")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("failed to read defaults for domain %s: %w", domain, err)
	}

	data := make(map[string]any)

	for line := range strings.SplitSeq(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " = ", 2)
		if len(parts) == 2 {
			key := strings.Trim(parts[0], "\"")
			value := strings.Trim(parts[1], "\";")
			data[key] = value
		}
	}

	return data, nil
}

func (bm *BackupManager) restoreDomainDefaults(ctx context.Context, domain string, data map[string]any) error {
	cmd := exec.CommandContext(ctx, "defaults", "delete", domain)
	if err := cmd.Run(); err != nil {
		bm.logger.Debug("Domain deletion failed (expected if domain doesn't exist)", "domain", domain)
	}

	for key, value := range data {
		valueStr := fmt.Sprintf("%v", value)
		cmd := exec.CommandContext(ctx, "defaults", "write", domain, key, valueStr)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restore key %s in domain %s: %w", key, domain, err)
		}
	}

	return nil
}

func (bm *BackupManager) saveBackupMetadata(backup *SystemBackup) error {
	metadataPath := filepath.Join(backup.BackupPath, "metadata.json")

	content := fmt.Sprintf(`{
  "backup_id": "%s",
  "timestamp": "%s",
  "domains": %v
}`, backup.BackupID, backup.Timestamp.Format(time.RFC3339), getKeys(backup.DomainData))

	return os.WriteFile(metadataPath, []byte(content), 0644)
}

func (bm *BackupManager) CleanupOldBackups() error {
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(entries) <= bm.maxBackups {
		return nil
	}

	type backupInfo struct {
		name    string
		modTime time.Time
	}

	var backups []backupInfo
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "backup_") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			backups = append(backups, backupInfo{
				name:    entry.Name(),
				modTime: info.ModTime(),
			})
		}
	}

	if len(backups) <= bm.maxBackups {
		return nil
	}

	for i := 0; i < len(backups)-bm.maxBackups; i++ {
		oldestPath := filepath.Join(bm.backupDir, backups[i].name)
		if err := os.RemoveAll(oldestPath); err != nil {
			bm.logger.Warn("Failed to remove old backup", "path", oldestPath, "error", err)
		} else {
			bm.logger.Debug("Removed old backup", "path", oldestPath)
		}
	}

	return nil
}

func getKeys(m map[string]map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

type TestRunner struct {
	backupManager *BackupManager
	logger        *logger.Logger
	testBackup    *SystemBackup
}

func NewTestRunner(logger *logger.Logger) *TestRunner {
	return &TestRunner{
		backupManager: NewBackupManager(logger, ""),
		logger:        logger.WithComponent("test-runner"),
	}
}

func (tr *TestRunner) SetupTest(ctx context.Context, domains []string) error {
	backup, err := tr.backupManager.CreateBackup(ctx, domains)
	if err != nil {
		return fmt.Errorf("failed to create test backup: %w", err)
	}

	tr.testBackup = backup
	tr.logger.Info("Test environment setup complete", "backup_id", backup.BackupID)
	return nil
}

func (tr *TestRunner) TeardownTest(ctx context.Context) error {
	if tr.testBackup == nil {
		return nil
	}

	if err := tr.backupManager.RestoreBackup(ctx, tr.testBackup); err != nil {
		return fmt.Errorf("failed to restore test backup: %w", err)
	}

	tr.logger.Info("Test environment restored", "backup_id", tr.testBackup.BackupID)
	tr.testBackup = nil
	return nil
}

func (tr *TestRunner) RunWithRollback(ctx context.Context, domains []string, testFunc func() error) error {
	if err := tr.SetupTest(ctx, domains); err != nil {
		return err
	}

	defer func() {
		if rollbackErr := tr.TeardownTest(ctx); rollbackErr != nil {
			tr.logger.Error("Failed to rollback test changes", "error", rollbackErr)
		}
	}()

	return testFunc()
}
