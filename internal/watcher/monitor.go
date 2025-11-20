package watcher

import (
	"os"
	"path/filepath"
	"time"
)

// monitorDirectories ensures required directories exist
// This prevents directories from disappearing due to Docker/Samba cleanup
func (w *Watcher) monitorDirectories() {
	defer w.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Ensure directories exist immediately on start
	w.ensureDirectories()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.ensureDirectories()
		}
	}
}

// ensureDirectories creates missing directories
func (w *Watcher) ensureDirectories() {
	dirs := []string{
		filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Processing),
		filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Processed),
		filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Failed),
		filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Ignored),
		filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Tmp),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			w.cfg.Logger.Warn("Directory missing, recreating", "path", dir)
			if err := os.MkdirAll(dir, 0755); err != nil {
				w.cfg.Logger.Error("Failed to create directory", "path", dir, "error", err)
			} else {
				w.cfg.Logger.Info("Directory recreated", "path", dir)
			}
		}
	}
}
