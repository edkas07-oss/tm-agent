package spool

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// PruneReport contains statistics of the retention pruning cycle.
type PruneReport struct {
	StaleTmpPruned  int
	StaleJsonPruned int
	QuotaPruned     int
	TotalRemaining  int
}

// PruneSpool cleans stale temporary files, expires old json files, and enforces FIFO quota.
func PruneSpool(spoolDir string, maxAgeHours, maxFiles, staleTmpMinutes int) (*PruneReport, error) {
	report := &PruneReport{}

	if _, err := os.Stat(spoolDir); os.IsNotExist(err) {
		return report, nil
	}

	now := time.Now()

	// 1. Clean stale .tmp files older than staleTmpMinutes
	staleTmpCutoff := now.Add(-time.Duration(staleTmpMinutes) * time.Minute)
	tmpEntries, err := filepath.Glob(filepath.Join(spoolDir, "*.tmp"))
	if err == nil {
		for _, f := range tmpEntries {
			fi, err := os.Stat(f)
			if err == nil && fi.ModTime().Before(staleTmpCutoff) {
				if os.Remove(f) == nil {
					report.StaleTmpPruned++
				}
			}
		}
	}

	// 2. Clean stale .json files older than maxAgeHours
	staleJsonCutoff := now.Add(-time.Duration(maxAgeHours) * time.Hour)
	jsonEntries, err := filepath.Glob(filepath.Join(spoolDir, "*.json"))
	if err == nil {
		for _, f := range jsonEntries {
			fi, err := os.Stat(f)
			if err == nil && fi.ModTime().Before(staleJsonCutoff) {
				if os.Remove(f) == nil {
					report.StaleJsonPruned++
				}
			}
		}
	}

	// 3. Enforce FIFO quota if remaining .json files exceed maxFiles
	activeEntries, err := filepath.Glob(filepath.Join(spoolDir, "*.json"))
	if err != nil {
		return report, fmt.Errorf("failed to list spool files: %w", err)
	}

	// Sort filenames lexicographically (chronological order given nanosecond prefix)
	sort.Strings(activeEntries)

	if maxFiles > 0 && len(activeEntries) > maxFiles {
		excess := len(activeEntries) - maxFiles
		for i := 0; i < excess; i++ {
			if os.Remove(activeEntries[i]) == nil {
				report.QuotaPruned++
			}
		}
		report.TotalRemaining = maxFiles
	} else {
		report.TotalRemaining = len(activeEntries)
	}

	return report, nil
}

// ListSpoolRecords returns all valid .json files in the spool directory, ordered chronologically.
func ListSpoolRecords(spoolDir string) ([]string, error) {
	if _, err := os.Stat(spoolDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(spoolDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read spool directory %s: %w", spoolDir, err)
	}

	var jsonFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			jsonFiles = append(jsonFiles, filepath.Join(spoolDir, e.Name()))
		}
	}

	sort.Strings(jsonFiles)
	return jsonFiles, nil
}
