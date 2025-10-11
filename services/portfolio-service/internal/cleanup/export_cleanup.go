package cleanup

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"
)

type ExportCleanup struct {
	exportDir string
	ttl       time.Duration
	interval  time.Duration
}

func NewExportCleanup(exportDir string, ttl, interval time.Duration) *ExportCleanup {
	return &ExportCleanup{
		exportDir: exportDir,
		ttl:       ttl,
		interval:  interval,
	}
}

func (ec *ExportCleanup) Start(ctx context.Context) {
	ticker := time.NewTicker(ec.interval)
	defer ticker.Stop()

	log.Printf("Export cleanup job started: dir=%s, ttl=%v, interval=%v", ec.exportDir, ec.ttl, ec.interval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Export cleanup job stopped")
			return
		case <-ticker.C:
			if err := ec.cleanup(); err != nil {
				log.Printf("Export cleanup error: %v", err)
			}
		}
	}
}

func (ec *ExportCleanup) cleanup() error {
	if _, err := os.Stat(ec.exportDir); os.IsNotExist(err) {
		return nil
	}

	now := time.Now()
	deleted := 0
	errors := 0

	err := filepath.Walk(ec.exportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		age := now.Sub(info.ModTime())
		if age > ec.ttl {
			if err := os.Remove(path); err != nil {
				log.Printf("Failed to delete expired export file %s: %v", path, err)
				errors++
				return nil
			}
			log.Printf("Deleted expired export file: %s (age: %v)", path, age)
			deleted++
		}

		return nil
	})

	if err != nil {
		return err
	}

	if deleted > 0 || errors > 0 {
		log.Printf("Export cleanup completed: deleted=%d, errors=%d", deleted, errors)
	}

	return nil
}
