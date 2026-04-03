package mock_test

import (
	"context"
	"testing"

	hivemind "github.com/smartify-inc/hivemind-go-sdk"
	"github.com/smartify-inc/hivemind-go-sdk/mock"
)

func TestMockClientReturnsConfiguredResponse(t *testing.T) {
	want := &hivemind.RetrieveResponse{
		ContextText:   "custom context",
		ContextTokens: 99,
	}
	c := mock.NewClient(mock.WithRetrieveResponse(want))

	got, err := c.Retrieve(context.Background(), &hivemind.RetrieveRequest{Intent: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ContextText != want.ContextText {
		t.Errorf("ContextText = %q, want %q", got.ContextText, want.ContextText)
	}
	if got.ContextTokens != want.ContextTokens {
		t.Errorf("ContextTokens = %d, want %d", got.ContextTokens, want.ContextTokens)
	}
}

func TestMockClientRecordsCalls(t *testing.T) {
	c := mock.NewClient()

	_, _ = c.Retrieve(context.Background(), &hivemind.RetrieveRequest{Intent: "lookup"})

	calls := c.CallsFor("Retrieve")
	if len(calls) != 1 {
		t.Fatalf("expected 1 Retrieve call, got %d", len(calls))
	}
	if calls[0].Method != "Retrieve" {
		t.Errorf("method = %q, want %q", calls[0].Method, "Retrieve")
	}
}

func TestMockClientDefaultResponses(t *testing.T) {
	c := mock.NewClient()

	resp, err := c.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "healthy" {
		t.Errorf("Status = %q, want %q", resp.Status, "healthy")
	}
}

func TestMockClientCallCount(t *testing.T) {
	c := mock.NewClient()

	_, _ = c.HealthCheck(context.Background())
	_, _ = c.Retrieve(context.Background(), &hivemind.RetrieveRequest{Intent: "a"})
	_, _ = c.Retrieve(context.Background(), &hivemind.RetrieveRequest{Intent: "b"})

	if n := c.CallCount(); n != 3 {
		t.Errorf("CallCount() = %d, want 3", n)
	}
}
