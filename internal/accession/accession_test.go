package accession

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateFileIDs(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "testfile.txt")

	// File should be created
	_, err := createFileIDFile(filePath, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check that the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("expected file to exist at %s", filePath)
	}

	// Calling again should return ErrFileAlreadyExists
	_, err = createFileIDFile(filePath, false)
	if err != ErrFileAlreadyExists {
		t.Fatalf("expected ErrFileAlreadyExists, got %v", err)
	}

	// Dry run should not create a file
	dryRunFile := filepath.Join(tmpDir, "dryrun.txt")
	_, err = createFileIDFile(dryRunFile, true)
	if err != nil {
		t.Fatalf("expected no error in dry run, got %v", err)
	}

	// Ensure file was NOT created in dry run
	if _, err := os.Stat(dryRunFile); !os.IsNotExist(err) {
		t.Fatalf("expected file NOT to exist in dry run, but it does")
	}
}
