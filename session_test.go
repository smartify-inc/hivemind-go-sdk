package hivemind_test

import (
	"context"
	"testing"

	hivemind "github.com/smartify-inc/hivemind-go-sdk"
	"github.com/smartify-inc/hivemind-go-sdk/mock"
)

func TestNewSessionDefaults(t *testing.T) {
	mc := mock.NewClient()
	s := hivemind.NewSession(mc,
		hivemind.WithWorkflowID("wf-1"),
		hivemind.WithTask("test task"),
	)

	if s.RunID() == "" {
		t.Fatal("expected auto-generated run ID")
	}
	if s.WorkflowID() != "wf-1" {
		t.Errorf("WorkflowID() = %q, want %q", s.WorkflowID(), "wf-1")
	}
	if s.Turn() != 0 {
		t.Errorf("Turn() = %d, want 0", s.Turn())
	}
}

func TestNewSessionCustomRunID(t *testing.T) {
	mc := mock.NewClient()
	s := hivemind.NewSession(mc,
		hivemind.WithWorkflowID("wf-1"),
		hivemind.WithTask("test"),
		hivemind.WithRunID("custom-run"),
	)
	if s.RunID() != "custom-run" {
		t.Errorf("RunID() = %q, want %q", s.RunID(), "custom-run")
	}
}

func TestSessionLifecycle(t *testing.T) {
	mc := mock.NewClient()
	ctx := context.Background()

	s := hivemind.NewSession(mc,
		hivemind.WithWorkflowID("wf-1"),
		hivemind.WithTask("lifecycle test"),
		hivemind.WithRunID("run-lc"),
	)

	ctxStr, err := s.GetContext(ctx, 4000)
	if err != nil {
		t.Fatalf("GetContext: %v", err)
	}
	if ctxStr != "mock context" {
		t.Errorf("GetContext = %q, want %q", ctxStr, "mock context")
	}
	if mc.CallCount() != 2 { // StartSession + GetContextWindow
		t.Errorf("call count = %d, want 2", mc.CallCount())
	}

	if err := s.RecordResponse(ctx, "hello world", 50); err != nil {
		t.Fatalf("RecordResponse: %v", err)
	}
	if s.Turn() != 1 {
		t.Errorf("Turn() = %d, want 1", s.Turn())
	}

	resp, err := s.End(ctx)
	if err != nil {
		t.Fatalf("End: %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("End status = %q, want %q", resp.Status, "completed")
	}

	if len(mc.CallsFor("RecordEpisode")) != 1 {
		t.Errorf("expected 1 RecordEpisode call, got %d", len(mc.CallsFor("RecordEpisode")))
	}
	if len(mc.CallsFor("EndSession")) != 1 {
		t.Errorf("expected 1 EndSession call, got %d", len(mc.CallsFor("EndSession")))
	}
}

func TestSessionStartsOnlyOnce(t *testing.T) {
	mc := mock.NewClient()
	ctx := context.Background()

	s := hivemind.NewSession(mc,
		hivemind.WithWorkflowID("wf-1"),
		hivemind.WithTask("once"),
		hivemind.WithRunID("run-once"),
	)

	_, _ = s.GetContext(ctx, 4000)
	_, _ = s.GetContext(ctx, 4000)

	if len(mc.CallsFor("StartSession")) != 1 {
		t.Errorf("StartSession called %d times, want 1", len(mc.CallsFor("StartSession")))
	}
}

func TestSessionEndWithoutStart(t *testing.T) {
	mc := mock.NewClient()
	ctx := context.Background()

	s := hivemind.NewSession(mc,
		hivemind.WithWorkflowID("wf-1"),
		hivemind.WithTask("no-start"),
		hivemind.WithRunID("run-ns"),
	)

	resp, err := s.End(ctx)
	if err != nil {
		t.Fatalf("End: %v", err)
	}
	if resp.Status != "not_started" {
		t.Errorf("status = %q, want %q", resp.Status, "not_started")
	}
	if mc.CallCount() != 0 {
		t.Errorf("expected no API calls, got %d", mc.CallCount())
	}
}

func TestSessionMultiTurn(t *testing.T) {
	mc := mock.NewClient()
	ctx := context.Background()

	s := hivemind.NewSession(mc,
		hivemind.WithWorkflowID("wf-1"),
		hivemind.WithTask("multi-turn"),
		hivemind.WithRunID("run-mt"),
	)

	_, _ = s.GetContext(ctx, 4000)
	_ = s.RecordResponse(ctx, "first response", 10)
	_, _ = s.GetContext(ctx, 4000)
	_ = s.RecordResponse(ctx, "second response", 20)

	if s.Turn() != 2 {
		t.Errorf("Turn() = %d, want 2", s.Turn())
	}

	_, _ = s.End(ctx)

	episodeCalls := mc.CallsFor("RecordEpisode")
	if len(episodeCalls) != 1 {
		t.Fatalf("expected 1 RecordEpisode, got %d", len(episodeCalls))
	}
}

func TestSessionAddObservation(t *testing.T) {
	mc := mock.NewClient()
	ctx := context.Background()

	s := hivemind.NewSession(mc,
		hivemind.WithWorkflowID("wf-1"),
		hivemind.WithTask("obs"),
		hivemind.WithRunID("run-obs"),
	)

	_, _ = s.GetContext(ctx, 4000)
	if err := s.AddObservation(ctx, "User seems frustrated"); err != nil {
		t.Fatalf("AddObservation: %v", err)
	}

	calls := mc.CallsFor("UpdateContext")
	if len(calls) != 1 {
		t.Fatalf("expected 1 UpdateContext, got %d", len(calls))
	}
}

func TestSessionAddToolResult(t *testing.T) {
	mc := mock.NewClient()
	ctx := context.Background()

	s := hivemind.NewSession(mc,
		hivemind.WithWorkflowID("wf-1"),
		hivemind.WithTask("tool"),
		hivemind.WithRunID("run-tool"),
	)

	_, _ = s.GetContext(ctx, 4000)
	if err := s.AddToolResult(ctx, "search", "3 results found"); err != nil {
		t.Fatalf("AddToolResult: %v", err)
	}

	if s.Turn() != 1 {
		t.Errorf("Turn() = %d, want 1", s.Turn())
	}
}
