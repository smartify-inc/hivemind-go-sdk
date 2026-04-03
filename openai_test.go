//go:build openai

package hivemind_test

import (
	"context"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	hivemind "github.com/smartify-inc/hivemind-go-sdk"
	"github.com/smartify-inc/hivemind-go-sdk/mock"
)

func testOpenAIClient(t *testing.T) openai.Client {
	t.Helper()
	return openai.NewClient(
		option.WithAPIKey("sk-test-dummy"),
		option.WithBaseURL("https://example.invalid/v1"),
	)
}

func TestWrapOpenAIConstruction(t *testing.T) {
	mc := mock.NewClient()
	oaiClient := testOpenAIClient(t)

	wrapped := hivemind.WrapOpenAI(oaiClient, mc,
		hivemind.WithWorkflowID("wf-1"),
		hivemind.WithTask("test"),
		hivemind.WithRunID("run-oai"),
	)

	if wrapped.Session().RunID() != "run-oai" {
		t.Errorf("RunID = %q, want %q", wrapped.Session().RunID(), "run-oai")
	}
}

func TestWrapOpenAIEndWithoutStart(t *testing.T) {
	mc := mock.NewClient()
	oaiClient := testOpenAIClient(t)
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
