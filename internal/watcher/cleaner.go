package watcher

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fabyo/gordon-watcher/internal/logger"
	"github.com/fabyo/gordon-watcher/internal/metrics"
)

// Cleaner removes empty directories
type Cleaner struct {
	workingDir    string
	protectedDirs map[string]bool
	interval      time.Duration
	logger        *logger.Logger
	stop          chan struct{}
	done          chan struct{}
}

// NewCleaner creates a new directory cleaner
func NewCleaner(workingDir string, protectedDirs []string, interval time.Duration, log *logger.Logger) *Cleaner {
	protected := make(map[string]bool)
	for _, dir := range protectedDirs {
		protected[dir] = true
	}

	return &Cleaner{
		workingDir:    workingDir,
		protectedDirs: protected,
		interval:      interval,
		logger:        log,
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
	}
}

// Start starts the cleaner
func (c *Cleaner) Start() {
	go c.run()
}

// Stop stops the cleaner
func (c *Cleaner) Stop() {
	close(c.stop)
	<-c.done
}

// run is the main cleanup loop
func (c *Cleaner) run() {
	defer close(c.done)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.cleanup()
		}
	}
}

// cleanup removes empty directories
func (c *Cleaner) cleanup() {
	c.logger.Debug("Starting directory cleanup")

	err := filepath.Walk(c.workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		if !info.IsDir() {
			return nil
		}

		// Don't remove the working directory itself
		if path == c.workingDir {
			return nil
		}

		// Don't remove protected directories
		if c.protectedDirs[path] {
			return nil
		}

		// Check if directory is empty
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		if len(entries) == 0 {
			c.logger.Debug("Removing empty directory", "path", path)

			if err := os.Remove(path); err != nil {
				c.logger.Error("Failed to remove empty directory",
					"path", path,
					"error", err)
			} else {
				metrics.EmptyDirectoriesRemoved.Inc()
			}
		}

		return nil
	})

	if err != nil {
		c.logger.Error("Error during cleanup", "error", err)
	}
}
