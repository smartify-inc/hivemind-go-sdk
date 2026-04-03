//go:build integration

package hivemind_test

import (
	"context"
	"os"
	"testing"
	"time"

	hivemind "github.com/smartify-inc/hivemind-go-sdk"
)

func getEndpoint() string {
	if ep := os.Getenv("HIVEMIND_GRPC_ENDPOINT"); ep != "" {
		return ep
	}
	return "localhost:50051"
}

func getRESTEndpoint() string {
	if ep := os.Getenv("HIVEMIND_REST_ENDPOINT"); ep != "" {
		return ep
	}
	return "http://localhost:8000"
}

func TestIntegrationSessionLifecycle(t *testing.T) {
	client, err := hivemind.NewClient(
		"sk_test_integration",
		hivemind.WithEndpoint(getEndpoint()),
		hivemind.WithTransport(hivemind.TransportGRPC),
		hivemind.WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 1. Start session
	startResp, err := client.StartSession(ctx, &hivemind.StartSessionRequest{
		RunID:      "integration-run-1",
		WorkflowID: "integration-wf",
		Task:       "Integration test task",
	})
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if startResp.RunID != "integration-run-1" {
		t.Errorf("expected run_id integration-run-1, got %s", startResp.RunID)
	}
	if startResp.Status != "active" {
		t.Errorf("expected status active, got %s", startResp.Status)
	}

	// 2. Update context 3 times
	for i := range 3 {
		err = client.UpdateContext(ctx, &hivemind.UpdateContextRequest{
			RunID: "integration-run-1",
			Field: "observations",
			Data:  map[string]any{"observation": i},
		})
		if err != nil {
			t.Fatalf("UpdateContext %d: %v", i, err)
		}
	}

	// 3. Get context window (full)
	ctxResp, err := client.GetContextWindow(ctx, &hivemind.ContextWindowRequest{
		RunID:          "integration-run-1",
		AgentID:        "test-agent",
		MaxTokens:      4000,
		ConversationID: "conv-1",
	})
	if err != nil {
		t.Fatalf("GetContextWindow (full): %v", err)
	}
	if ctxResp.ModeUsed != "full" {
		t.Errorf("expected mode full, got %s", ctxResp.ModeUsed)
	}
	if ctxResp.TotalTokens <= 0 {
		t.Errorf("expected tokens > 0, got %d", ctxResp.TotalTokens)
	}

	// 4. Update and get delta
	err = client.UpdateContext(ctx, &hivemind.UpdateContextRequest{
		RunID: "integration-run-1",
		Field: "observations",
		Data:  "new observation after full pack",
	})
	if err != nil {
		t.Fatalf("UpdateContext for delta: %v", err)
	}

	ctxResp2, err := client.GetContextWindow(ctx, &hivemind.ContextWindowRequest{
		RunID:          "integration-run-1",
		AgentID:        "test-agent",
		MaxTokens:      4000,
		ConversationID: "conv-1",
	})
	if err != nil {
		t.Fatalf("GetContextWindow (delta): %v", err)
	}
	if ctxResp2.ModeUsed != "delta" {
		t.Logf("expected delta mode, got %s (may be full if context is small)", ctxResp2.ModeUsed)
	}

	// 5. End session
	endResp, err := client.EndSession(ctx, "integration-run-1")
	if err != nil {
		t.Fatalf("EndSession: %v", err)
	}
	if endResp.Status != "completed" {
		t.Errorf("expected status completed, got %s", endResp.Status)
	}
}

func TestIntegrationRecordAndRetrieve(t *testing.T) {
	client, err := hivemind.NewClient(
		"sk_test_integration",
		hivemind.WithEndpoint(getEndpoint()),
		hivemind.WithTransport(hivemind.TransportGRPC),
		hivemind.WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Record an episode
	recResp, err := client.RecordEpisode(ctx, &hivemind.RecordEpisodeRequest{
		WorkflowID:      "integration-wf",
		RunID:           "integration-rec-1",
		TaskDescription: "Write a blog post about testing",
		StepResults: map[string]hivemind.StepResult{
			"research": {Action: "Researched topic", Outcome: "success", TokensUsed: 500},
			"draft":    {Action: "Wrote draft", Outcome: "success", TokensUsed: 1500},
		},
		TotalTokens: 2000,
		DurationMs:  5000,
	})
	if err != nil {
		t.Fatalf("RecordEpisode: %v", err)
	}
	if recResp.EpisodeID == "" {
		t.Error("expected non-empty episode_id")
	}

	// Retrieve
	retResp, err := client.Retrieve(ctx, &hivemind.RetrieveRequest{
		Intent:    "Write a blog post",
		MaxTokens: 2000,
	})
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if retResp.ContextTokens < 0 {
		t.Errorf("expected context_tokens >= 0, got %d", retResp.ContextTokens)
	}
}

func TestIntegrationStatsAndHealth(t *testing.T) {
	client, err := hivemind.NewClient(
		"sk_test_integration",
		hivemind.WithEndpoint(getEndpoint()),
		hivemind.WithTransport(hivemind.TransportGRPC),
		hivemind.WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Health check
	healthResp, err := client.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	if healthResp.Status != "healthy" {
		t.Errorf("expected healthy, got %s", healthResp.Status)
	}

	// Stats
	statsResp, err := client.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if statsResp.TotalEpisodes < 0 {
		t.Errorf("expected total_episodes >= 0, got %d", statsResp.TotalEpisodes)
	}
}

func TestIntegrationRESTTransport(t *testing.T) {
	client, err := hivemind.NewClient(
		"sk_test_integration",
		hivemind.WithEndpoint(getRESTEndpoint()),
		hivemind.WithTransport(hivemind.TransportREST),
		hivemind.WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient REST: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	healthResp, err := client.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck REST: %v", err)
	}
	if healthResp.Status != "healthy" {
		t.Errorf("expected healthy, got %s", healthResp.Status)
	}
}
