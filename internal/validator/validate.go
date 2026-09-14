package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eddywiyatno/tm-agent/pkg/termutil"
)

var requiredContractFiles = []string{
	"CONFIG",
	"PROJECT",
	"VERSION",
	"AGENTS.md",
	"README.md",
}

var forbiddenExtensions = []string{
	".pem",
	".key",
	".p12",
	".pfx",
	".pkcs12",
}

var forbiddenFilenames = []string{
	"id_rsa",
	"id_ecdsa",
	"id_ed25519",
	".env",
	".env.local",
	".env.production",
}

// ValidateRepository performs contract, layout, and secret leakage audits.
func ValidateRepository(rootDir string) error {
	termutil.PrintInfo("Starting baseline validation on project root: %s", rootDir)

	// Step 1: Layout validation
	fmt.Println("[1/3] Validating repository layout and required contract files...")
	for _, req := range requiredContractFiles {
		p := filepath.Join(rootDir, req)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return fmt.Errorf("missing required contract file: %s", req)
		}
	}

	// Step 2: Secret leakage audit
	fmt.Println("[2/3] Auditing repository for forbidden sensitive material files...")
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "bin" || name == "dist" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		base := info.Name()
		for _, forbid := range forbiddenFilenames {
			if base == forbid {
				return fmt.Errorf("forbidden sensitive file detected: %s", path)
			}
		}

		ext := strings.ToLower(filepath.Ext(base))
		for _, forbidExt := range forbiddenExtensions {
			if ext == forbidExt {
				return fmt.Errorf("forbidden certificate/key extension detected (%s): %s", ext, path)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Step 3: Schema & JSON syntax validation
	fmt.Println("[3/3] Validating JSON schema syntax integrity across configuration files...")
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "bin" {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(info.Name(), ".json") {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return fmt.Errorf("failed to read json file %s: %w", path, readErr)
			}
			var parsed interface{}
			if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
				return fmt.Errorf("invalid json syntax in %s: %w", path, jsonErr)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	termutil.PrintSuccess("All platform validation assertions passed successfully.")
	return nil
}
