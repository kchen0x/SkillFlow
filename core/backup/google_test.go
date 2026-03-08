package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoogleCredentialOptionAcceptsInlineJSON(t *testing.T) {
	opt, err := googleCredentialOption(`{"type":"service_account","project_id":"demo"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opt == nil {
		t.Fatal("expected client option, got nil")
	}
}

func TestGoogleCredentialOptionAcceptsFilePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "service-account.json")
	if err := os.WriteFile(path, []byte(`{"type":"service_account","project_id":"demo"}`), 0644); err != nil {
		t.Fatalf("write temp credentials failed: %v", err)
	}

	opt, err := googleCredentialOption(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opt == nil {
		t.Fatal("expected client option, got nil")
	}
}

func TestGoogleCredentialOptionRejectsEmptyValue(t *testing.T) {
	if _, err := googleCredentialOption("   "); err == nil {
		t.Fatal("expected error for empty credential input")
	}
}
