package remarkable

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrXochitlNotFound = errors.New("xochitl process not found - is the reMarkable software running?")

func findXochitlPID() (string, error) {
	return findXochitlPIDIn("/proc")
}

// findXochitlPIDIn scans a procfs root for the process whose executable is
// /usr/bin/xochitl. Errors on individual processes are skipped rather than
// reported: a process exiting between the listing of base and the reading of
// its directory is routine, and must not abort the scan.
func findXochitlPIDIn(base string) (string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", base, err)
	}

	for _, entry := range entries {
		pid := entry.Name()
		if !entry.IsDir() {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(base, pid))
		if err != nil {
			continue
		}
		for _, entry := range entries {
			// Type() reads the d_type returned by readdir, so unlike Info() it
			// costs no extra stat per entry.
			if entry.Type()&os.ModeSymlink == 0 {
				continue
			}
			orig, err := os.Readlink(filepath.Join(base, pid, entry.Name()))
			if err != nil {
				continue
			}
			if orig == "/usr/bin/xochitl" {
				return pid, nil
			}
		}
	}
	return "", ErrXochitlNotFound
}
