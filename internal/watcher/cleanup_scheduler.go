package watcher

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fabyo/gordon-watcher/internal/logger"
	"github.com/robfig/cron/v3"
)

// CleanupScheduler handles scheduled directory cleanup
type CleanupScheduler struct {
	workingDir string
	subDirs    SubDirectories
	retention  map[string]int
	schedule   string
	cron       *cron.Cron
	logger     *logger.Logger
	isRunning  bool
}

// CleanupConfig holds cleanup configuration
type CleanupConfig struct {
	WorkingDir string
	SubDirs    SubDirectories
	Retention  map[string]int
	Schedule   string
	Logger     *logger.Logger
}

// NewCleanupScheduler creates a new cleanup scheduler
func NewCleanupScheduler(cfg CleanupConfig) *CleanupScheduler {
	return &CleanupScheduler{
		workingDir: cfg.WorkingDir,
		subDirs:    cfg.SubDirs,
		retention:  cfg.Retention,
		schedule:   cfg.Schedule,
		logger:     cfg.Logger,
		cron:       cron.New(),
	}
}

// Start starts the cleanup scheduler
func (cs *CleanupScheduler) Start() error {
	// Add cleanup job based on schedule
	_, err := cs.cron.AddFunc(cs.schedule, func() {
		cs.runCleanup()
	})
	if err != nil {
		return err
	}

	cs.cron.Start()
	cs.isRunning = true
	cs.logger.Info("Cleanup scheduler started", "schedule", cs.schedule)

	return nil
}

// Stop stops the cleanup scheduler
func (cs *CleanupScheduler) Stop() {
	if cs.isRunning {
		cs.cron.Stop()
		cs.isRunning = false
		cs.logger.Info("Cleanup scheduler stopped")
	}
}

// runCleanup performs the cleanup based on retention policies
func (cs *CleanupScheduler) runCleanup() {
	cs.logger.Info("Starting scheduled cleanup")

	// Clean tmp directory (always clean immediately)
	cs.cleanDirectory(filepath.Join(cs.workingDir, cs.subDirs.Tmp), 0, "tmp")

	// Clean processed directory (if retention > 0)
	if retention, ok := cs.retention["processed"]; ok && retention > 0 {
		cs.cleanDirectory(filepath.Join(cs.workingDir, cs.subDirs.Processed), retention, "processed")
	}

	// Clean failed directory
	if retention, ok := cs.retention["failed"]; ok && retention > 0 {
		cs.cleanDirectory(filepath.Join(cs.workingDir, cs.subDirs.Failed), retention, "failed")
	}

	// Clean ignored directory
	if retention, ok := cs.retention["ignored"]; ok && retention > 0 {
		cs.cleanDirectory(filepath.Join(cs.workingDir, cs.subDirs.Ignored), retention, "ignored")
	}

	cs.logger.Info("Scheduled cleanup completed")
}

// cleanDirectory removes files older than retention days
func (cs *CleanupScheduler) cleanDirectory(dir string, retentionDays int, dirType string) {
	now := time.Now()
	cutoff := now.AddDate(0, 0, -retentionDays)

	deletedCount := 0
	deletedSize := int64(0)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// For tmp (retention = 0), delete all files
		// For others, delete files older than cutoff
		shouldDelete := false
		if retentionDays == 0 {
			shouldDelete = true
		} else if info.ModTime().Before(cutoff) {
			shouldDelete = true
		}

		if shouldDelete {
			cs.logger.Info("Deleting file during cleanup", "path", path, "retention_days", retentionDays, "age_days", int(now.Sub(info.ModTime()).Hours()/24))
			if err := os.Remove(path); err != nil {
				cs.logger.Warn("Failed to delete file",
					"path", path,
					"error", err)
			} else {
				deletedCount++
				deletedSize += info.Size()
			}
		}

		return nil
	})

	if err != nil {
		cs.logger.Error("Failed to walk directory",
			"dir", dir,
			"error", err)
		return
	}

	if deletedCount > 0 {
		cs.logger.Info("Cleanup completed",
			"directory", dirType,
			"files_deleted", deletedCount,
			"bytes_freed", deletedSize,
			"retention_days", retentionDays)
	}
}
