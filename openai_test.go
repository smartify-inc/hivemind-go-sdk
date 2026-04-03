//go:build openai

package hivemind_test

import (
	"context"
	"testing"

	openai "github.com/sashabaranov/go-openai"

	hivemind "github.com/smartifyai/hivemind-go"
	"github.com/smartifyai/hivemind-go/mock"
)

// NOTE: These tests use the build tag "openai" so they only compile when
// the go-openai dependency is present. Since WrappedOpenAI.CreateChatCompletion
// calls the real OpenAI client internally, these tests verify construction
// and session wiring rather than full HTTP round-trips.

func TestWrapOpenAIConstruction(t *testing.T) {
	mc := mock.NewClient()
	oaiClient := openai.NewClient("sk-test")

	wrapped := hivemind.WrapOpenAI(oaiClient, mc,
		hivemind.WithWorkflowID("wf-1"),
		hivemind.WithTask("test"),
		hivemind.WithRunID("run-oai"),
	)

	if wrapped.Session().RunID() != "run-oai" {
		t.Errorf("RunID = %q, want %q", wrapped.Session().RunID(), "run-oai")
	}
	if wrapped.Inner() != oaiClient {
		t.Error("Inner() should return the original OpenAI client")
	}
}

func TestWrapOpenAIEndWithoutStart(t *testing.T) {
	mc := mock.NewClient()
	oaiClient := openai.NewClient("sk-test")
	ctx := context.Background()

	wrapped := hivemind.WrapOpenAI(oaiClient, mc,
		hivemind.WithWorkflowID("wf-1"),
		hivemind.WithTask("no-start"),
		hivemind.WithRunID("run-ns"),
	)

	resp, err := wrapped.End(ctx)
	if err != nil {
		t.Fatalf("End: %v", err)
	}
	if resp.Status != "not_started" {
		t.Errorf("status = %q, want %q", resp.Status, "not_started")
	}
}
