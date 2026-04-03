package hivemind_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	hivemind "github.com/smartifyai/hivemind-go"
)

func TestHivemindErrorMessage(t *testing.T) {
	err := &hivemind.HivemindError{
		Title:  "Not Found",
		Detail: "resource missing",
		Status: 404,
	}
	msg := err.Error()
	if !strings.Contains(msg, "Not Found") {
		t.Errorf("expected title in message, got %q", msg)
	}
	if !strings.Contains(msg, "resource missing") {
		t.Errorf("expected detail in message, got %q", msg)
	}
	if !strings.Contains(msg, "404") {
		t.Errorf("expected status in message, got %q", msg)
	}
}

func TestIsNotFound(t *testing.T) {
	err := &hivemind.HivemindError{Status: 404, Title: "Not Found"}
	if !hivemind.IsNotFound(err) {
		t.Error("IsNotFound should return true for 404")
	}
}

func TestIsUnauthorized(t *testing.T) {
	err := &hivemind.HivemindError{Status: 401, Title: "Unauthorized"}
	if !hivemind.IsUnauthorized(err) {
		t.Error("IsUnauthorized should return true for 401")
	}
}

func TestIsForbidden(t *testing.T) {
	err := &hivemind.HivemindError{Status: 403, Title: "Forbidden"}
	if !hivemind.IsForbidden(err) {
		t.Error("IsForbidden should return true for 403")
	}
}

func TestIsRateLimited(t *testing.T) {
	err := &hivemind.HivemindError{Status: 429, Title: "Too Many Requests"}
	if !hivemind.IsRateLimited(err) {
		t.Error("IsRateLimited should return true for 429")
	}
}

func TestIsConflict(t *testing.T) {
	err := &hivemind.HivemindError{Status: 409, Title: "Conflict"}
	if !hivemind.IsConflict(err) {
		t.Error("IsConflict should return true for 409")
	}
}

func TestIsServerError(t *testing.T) {
	for _, code := range []int{500, 502, 503} {
		err := &hivemind.HivemindError{Status: code, Title: "Server Error"}
		if !hivemind.IsServerError(err) {
			t.Errorf("IsServerError should return true for %d", code)
		}
	}
}

func TestSentinelFalseForWrongStatus(t *testing.T) {
	err := &hivemind.HivemindError{Status: 401, Title: "Unauthorized"}
	if hivemind.IsNotFound(err) {
		t.Error("IsNotFound should return false for 401")
	}
}

func TestErrorsAs(t *testing.T) {
	inner := &hivemind.HivemindError{Status: 404, Title: "Not Found", Detail: "gone"}
	wrapped := fmt.Errorf("operation failed: %w", inner)

	var he *hivemind.HivemindError
	if !errors.As(wrapped, &he) {
		t.Fatal("errors.As should unwrap *HivemindError")
	}
	if he.Status != 404 {
		t.Errorf("unwrapped status = %d, want 404", he.Status)
	}
}
