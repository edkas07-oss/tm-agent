package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRepositoryValid(t *testing.T) {
	tmpDir := t.TempDir()

	for _, req := range requiredContractFiles {
		_ = os.WriteFile(filepath.Join(tmpDir, req), []byte("test"), 0644)
	}

	if err := ValidateRepository(tmpDir); err != nil {
		t.Fatalf("Validation expected to pass, got: %v", err)
	}
}

func TestValidateRepositoryMissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tmpDir, "CONFIG"), []byte("test"), 0644)
	// Missing PROJECT, VERSION, AGENTS.md, README.md

	if err := ValidateRepository(tmpDir); err == nil {
		t.Errorf("Validation expected to fail for missing files, got nil")
	}
}

func TestValidateRepositoryForbiddenFile(t *testing.T) {
	tmpDir := t.TempDir()

	for _, req := range requiredContractFiles {
		_ = os.WriteFile(filepath.Join(tmpDir, req), []byte("test"), 0644)
	}

	// Add forbidden file
	_ = os.WriteFile(filepath.Join(tmpDir, "server.key"), []byte("secret"), 0644)

	if err := ValidateRepository(tmpDir); err == nil {
		t.Errorf("Validation expected to fail for forbidden .key file, got nil")
	}
}
