package remarkable

import (
	"os"
	"path/filepath"
	"testing"
)

// fakeProc builds a procfs-shaped tree: one directory per pid, each holding an
// "exe" symlink to the process executable.
func fakeProc(t *testing.T, exeByPID map[string]string) string {
	t.Helper()
	base := t.TempDir()
	for pid, exe := range exeByPID {
		dir := filepath.Join(base, pid)
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(exe, filepath.Join(dir, "exe")); err != nil {
			t.Fatal(err)
		}
	}
	return base
}

// TestFindXochitlPIDIn_Found checks the happy path against a synthetic procfs.
func TestFindXochitlPIDIn_Found(t *testing.T) {
	base := fakeProc(t, map[string]string{
		"1":    "/sbin/init",
		"1234": "/usr/bin/xochitl",
	})

	pid, err := findXochitlPIDIn(base)
	if err != nil {
		t.Fatalf("findXochitlPIDIn() error = %v", err)
	}
	if pid != "1234" {
		t.Errorf("pid = %q, want %q", pid, "1234")
	}
}

// TestFindXochitlPIDIn_NotFound checks that an absent xochitl is a normal
// ErrXochitlNotFound rather than a match on some other process.
func TestFindXochitlPIDIn_NotFound(t *testing.T) {
	base := fakeProc(t, map[string]string{"1": "/sbin/init"})

	if _, err := findXochitlPIDIn(base); err != ErrXochitlNotFound {
		t.Errorf("err = %v, want %v", err, ErrXochitlNotFound)
	}
}

// TestFindXochitlPIDIn_MissingProcRoot is the regression guard for the bug this
// fixes: an unreadable /proc used to call log.Fatal, killing the process from
// inside a function that already returns an error. It also made the package
// untestable anywhere without procfs.
func TestFindXochitlPIDIn_MissingProcRoot(t *testing.T) {
	_, err := findXochitlPIDIn(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected an error for a missing proc root")
	}
	if err == ErrXochitlNotFound {
		t.Error("a missing proc root should not be reported as 'xochitl not running'")
	}
}

// TestFindXochitlPIDIn_SkipsUnreadableProcess covers the real-hardware race: a
// process exiting between the listing of /proc and the reading of its directory
// must not abort the scan, let alone terminate the server.
func TestFindXochitlPIDIn_SkipsUnreadableProcess(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read a 0000 directory")
	}
	base := fakeProc(t, map[string]string{
		"1234": "/usr/bin/xochitl",
		"999":  "/usr/bin/other",
	})
	// Make 999 unreadable, standing in for a process that vanished.
	vanished := filepath.Join(base, "999")
	if err := os.Chmod(vanished, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(vanished, 0o755) })

	pid, err := findXochitlPIDIn(base)
	if err != nil {
		t.Fatalf("scan aborted on an unreadable process: %v", err)
	}
	if pid != "1234" {
		t.Errorf("pid = %q, want %q", pid, "1234")
	}
}

// TestFindXochitlPIDReturnsError tests that findXochitlPID returns an error
// when the xochitl process is not found, rather than an empty string.
func TestFindXochitlPIDReturnsError(t *testing.T) {
	// This test is mainly for documentation and will likely skip on non-reMarkable hardware
	// The important thing is that the function signature returns (string, error)

	pid, err := findXochitlPID()

	// If running on reMarkable with xochitl running
	if err == nil && pid != "" {
		// Valid case - process found
		t.Logf("Found xochitl process: %s", pid)
		return
	}

	// If not running on reMarkable or xochitl not running
	if err != nil && pid == "" {
		// Also valid - error returned with empty PID
		t.Logf("xochitl not found (expected): %v", err)
		return
	}

	// Invalid combinations
	if err == nil && pid == "" {
		t.Error("findXochitlPID() returned empty string without error")
	}
	if err != nil && pid != "" {
		t.Errorf("findXochitlPID() returned PID %s with error %v", pid, err)
	}
}

// TestFindXochitlPIDErrorMessage tests that the error message is descriptive.
func TestFindXochitlPIDErrorMessage(t *testing.T) {
	// Try to find xochitl
	pid, err := findXochitlPID()

	// If not found, check error message
	if err != nil {
		errMsg := err.Error()
		// Error message should mention xochitl
		if errMsg == "" {
			t.Error("Error message is empty")
		}
		t.Logf("Error message: %s", errMsg)
		return
	}

	// Process found
	t.Logf("xochitl process found: %s", pid)
}
