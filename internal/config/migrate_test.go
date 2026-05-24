package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestMoveLegacyToPrimary_HappyPath(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, LegacyDir, Filename)
	mustWrite(t, legacy, "version: 1\nproject:\n  name: demo\n")

	moved, err := MoveLegacyToPrimary(root)
	if err != nil {
		t.Fatalf("MoveLegacyToPrimary: %v", err)
	}
	if !moved {
		t.Fatalf("moved = false, want true")
	}

	if _, err := os.Stat(legacy); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("legacy file still present: %v", err)
	}
	primary := filepath.Join(root, PrimaryDir, Filename)
	body, err := os.ReadFile(primary)
	if err != nil {
		t.Fatalf("read primary: %v", err)
	}
	if string(body) != "version: 1\nproject:\n  name: demo\n" {
		t.Errorf("primary content mismatch: %q", body)
	}
}

func TestMoveLegacyToPrimary_LegacyDirIsRemovedWhenEmpty(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, LegacyDir, Filename)
	mustWrite(t, legacy, "version: 1\n")

	if _, err := MoveLegacyToPrimary(root); err != nil {
		t.Fatalf("MoveLegacyToPrimary: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, LegacyDir)); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("empty .ai/ dir should be removed: %v", err)
	}
}

func TestMoveLegacyToPrimary_LegacyDirRetainedWhenNotEmpty(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, LegacyDir, Filename), "version: 1\n")
	mustWrite(t, filepath.Join(root, LegacyDir, "other.txt"), "user file\n")

	if _, err := MoveLegacyToPrimary(root); err != nil {
		t.Fatalf("MoveLegacyToPrimary: %v", err)
	}
	info, err := os.Stat(filepath.Join(root, LegacyDir))
	if err != nil || !info.IsDir() {
		t.Errorf("legacy dir with user files must survive: info=%v err=%v", info, err)
	}
}

func TestMoveLegacyToPrimary_PrimaryAlreadyExists_NoOp(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, LegacyDir, Filename), "version: 1\n# legacy\n")
	mustWrite(t, filepath.Join(root, PrimaryDir, Filename), "version: 1\n# primary\n")

	moved, err := MoveLegacyToPrimary(root)
	if err != nil {
		t.Fatalf("MoveLegacyToPrimary: %v", err)
	}
	if moved {
		t.Errorf("moved = true, want false when primary exists")
	}
	// Both files must still be present.
	if _, err := os.Stat(filepath.Join(root, LegacyDir, Filename)); err != nil {
		t.Errorf("legacy should be untouched: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, PrimaryDir, Filename)); err != nil {
		t.Errorf("primary should be untouched: %v", err)
	}
}

func TestMoveLegacyToPrimary_LegacyMissing(t *testing.T) {
	root := t.TempDir()
	_, err := MoveLegacyToPrimary(root)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
}

func TestMoveLegacyToPrimary_Idempotent(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, LegacyDir, Filename), "version: 1\n")

	if moved, err := MoveLegacyToPrimary(root); err != nil || !moved {
		t.Fatalf("first call: moved=%v err=%v", moved, err)
	}
	moved, err := MoveLegacyToPrimary(root)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if moved {
		t.Errorf("second call moved = true; expected no-op")
	}
}
