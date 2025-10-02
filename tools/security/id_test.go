package security_test

import (
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/thewandererbg/pgbase/tools/security"
)

func TestGeneratePrefixedULID(t *testing.T) {
	result := security.GeneratePrefixedULID("user")

	if !strings.HasPrefix(result, "user_") {
		t.Errorf("expected prefix 'user_', got %q", result)
	}

	ulidPart := strings.TrimPrefix(result, "user_")
	if _, err := ulid.Parse(ulidPart); err != nil {
		t.Errorf("invalid ULID part: %v", err)
	}
}

func TestGenerateULID(t *testing.T) {
	result := security.GenerateULID()

	if len(result) != 26 {
		t.Errorf("expected length 26, got %d", len(result))
	}

	if _, err := ulid.Parse(result); err != nil {
		t.Errorf("invalid ULID: %v", err)
	}
}

func TestGenerateUUIDv4(t *testing.T) {
	result := security.GenerateUUIDv4()

	if len(result) != 36 {
		t.Errorf("expected length 36, got %d", len(result))
	}

	parts := strings.Split(result, "-")
	if len(parts) != 5 {
		t.Errorf("expected 5 parts separated by hyphens, got %d", len(parts))
	}

	if !strings.HasPrefix(parts[2], "4") {
		t.Errorf("expected version 4 UUID")
	}
}

func TestGenerateUUIDv7(t *testing.T) {
	result := security.GenerateUUIDv7()

	if len(result) != 36 {
		t.Errorf("expected length 36, got %d", len(result))
	}

	parts := strings.Split(result, "-")
	if len(parts) != 5 {
		t.Errorf("expected 5 parts separated by hyphens, got %d", len(parts))
	}

	if !strings.HasPrefix(parts[2], "7") {
		t.Errorf("expected version 7 UUID")
	}
}
