package mock

import (
	"context"
	"sync"
	"time"

	hivemind "github.com/smartify-inc/hivemind-go-sdk"
)

// MockCall records a single method invocation on the mock client.
type MockCall struct {
	Method string
	Args   []any
}

// MockClient is a test double that implements hivemind.Client.
// It records all calls and returns configured or default responses.
type MockClient struct {
	mu    sync.Mutex
	Calls []MockCall

	StartSessionResponse     *hivemind.StartSessionResponse
	ContextWindowResponse    *hivemind.ContextWindowResponse
	EndSessionResponse       *hivemind.EndSessionResponse
	RetrieveResponse         *hivemind.RetrieveResponse
	RecordEpisodeResponse    *hivemind.RecordEpisodeResponse
	StatsResponse            *hivemind.StatsResponse
	ConsolidationResponse    *hivemind.ConsolidationResponse
	HealthResponse           *hivemind.HealthResponse

	StartSessionError        error
	UpdateContextError       error
	ContextWindowError       error
	EndSessionError          error
	RetrieveError            error
	RecordEpisodeError       error
	RecordFeedbackError      error
	StatsError               error
	ConsolidationError       error
	HealthError              error
}

// MockOption configures a MockClient.
type MockOption func(*MockClient)

// NewClient creates a new MockClient with the given options.
func NewClient(opts ...MockOption) *MockClient {
	c := &MockClient{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *MockClient) record(method string, args ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Calls = append(c.Calls, MockCall{Method: method, Args: args})
}

// CallCount returns the total number of recorded calls.
func (c *MockClient) CallCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.Calls)
}

// CallsFor returns all recorded calls matching the given method name.
func (c *MockClient) CallsFor(method string) []MockCall {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []MockCall
	for _, call := range c.Calls {
		if call.Method == method {
			out = append(out, call)
		}
	}
	return out
}

// Reset clears all recorded calls.
func (c *MockClient) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Calls = nil
}

func (c *MockClient) StartSession(_ context.Context, req *hivemind.StartSessionRequest) (*hivemind.StartSessionResponse, error) {
	c.record("StartSession", req)
	if c.StartSessionError != nil {
		return nil, c.StartSessionError
	}
	if c.StartSessionResponse != nil {
		return c.StartSessionResponse, nil
	}
	return &hivemind.StartSessionResponse{
		RunID:     req.RunID,
		Status:    "active",
		CreatedAt: time.Now(),
	}, nil
}

func (c *MockClient) UpdateContext(_ context.Context, req *hivemind.UpdateContextRequest) error {
	c.record("UpdateContext", req)
	return c.UpdateContextError
}

func (c *MockClient) GetContextWindow(_ context.Context, req *hivemind.ContextWindowRequest) (*hivemind.ContextWindowResponse, error) {
	c.record("GetContextWindow", req)
	if c.ContextWindowError != nil {
		return nil, c.ContextWindowError
	}
	if c.ContextWindowResponse != nil {
		return c.ContextWindowResponse, nil
	}
	return &hivemind.ContextWindowResponse{
		Context:  "mock context",
		ModeUsed: "full",
	}, nil
}

func (c *MockClient) EndSession(_ context.Context, runID string) (*hivemind.EndSessionResponse, error) {
	c.record("EndSession", runID)
	if c.EndSessionError != nil {
		return nil, c.EndSessionError
	}
	if c.EndSessionResponse != nil {
		return c.EndSessionResponse, nil
	}
	return &hivemind.EndSessionResponse{
		RunID:  runID,
		Status: "completed",
	}, nil
}

func (c *MockClient) Retrieve(_ context.Context, req *hivemind.RetrieveRequest) (*hivemind.RetrieveResponse, error) {
	c.record("Retrieve", req)
	if c.RetrieveError != nil {
		return nil, c.RetrieveError
	}
	if c.RetrieveResponse != nil {
		return c.RetrieveResponse, nil
	}
	return &hivemind.RetrieveResponse{
		ContextText:   "mock retrieved context",
		ContextTokens: 42,
	}, nil
}

func (c *MockClient) RecordEpisode(_ context.Context, req *hivemind.RecordEpisodeRequest) (*hivemind.RecordEpisodeResponse, error) {
	c.record("RecordEpisode", req)
	if c.RecordEpisodeError != nil {
		return nil, c.RecordEpisodeError
	}
	if c.RecordEpisodeResponse != nil {
		return c.RecordEpisodeResponse, nil
	}
	return &hivemind.RecordEpisodeResponse{
		EpisodeID: "mock-episode-id",
		TaskType:  "general",
	}, nil
}

func (c *MockClient) RecordFeedback(_ context.Context, req *hivemind.RecordFeedbackRequest) error {
	c.record("RecordFeedback", req)
	return c.RecordFeedbackError
}

func (c *MockClient) GetStats(_ context.Context) (*hivemind.StatsResponse, error) {
	c.record("GetStats")
	if c.StatsError != nil {
		return nil, c.StatsError
	}
	if c.StatsResponse != nil {
		return c.StatsResponse, nil
	}
	return &hivemind.StatsResponse{}, nil
}

func (c *MockClient) TriggerConsolidation(_ context.Context) (*hivemind.ConsolidationResponse, error) {
	c.record("TriggerConsolidation")
	if c.ConsolidationError != nil {
		return nil, c.ConsolidationError
	}
	if c.ConsolidationResponse != nil {
		return c.ConsolidationResponse, nil
	}
	return &hivemind.ConsolidationResponse{
		ConsolidationID: "mock-consolidation-id",
		Status:          "queued",
	}, nil
}

func (c *MockClient) HealthCheck(_ context.Context) (*hivemind.HealthResponse, error) {
	c.record("HealthCheck")
	if c.HealthError != nil {
		return nil, c.HealthError
	}
	if c.HealthResponse != nil {
		return c.HealthResponse, nil
	}
	return &hivemind.HealthResponse{
		Status:  "healthy",
		Version: "mock",
	}, nil
}

func (c *MockClient) Close() error {
	c.record("Close")
	return nil
}

// --- Functional options ---

func WithStartSessionResponse(r *hivemind.StartSessionResponse) MockOption {
	return func(c *MockClient) { c.StartSessionResponse = r }
}

func WithContextWindowResponse(r *hivemind.ContextWindowResponse) MockOption {
	return func(c *MockClient) { c.ContextWindowResponse = r }
}

func WithEndSessionResponse(r *hivemind.EndSessionResponse) MockOption {
	return func(c *MockClient) { c.EndSessionResponse = r }
}

func WithRetrieveResponse(r *hivemind.RetrieveResponse) MockOption {
	return func(c *MockClient) { c.RetrieveResponse = r }
}

func WithRecordEpisodeResponse(r *hivemind.RecordEpisodeResponse) MockOption {
	return func(c *MockClient) { c.RecordEpisodeResponse = r }
}

func WithStatsResponse(r *hivemind.StatsResponse) MockOption {
	return func(c *MockClient) { c.StatsResponse = r }
}

func WithConsolidationResponse(r *hivemind.ConsolidationResponse) MockOption {
	return func(c *MockClient) { c.ConsolidationResponse = r }
}

func WithHealthResponse(r *hivemind.HealthResponse) MockOption {
	return func(c *MockClient) { c.HealthResponse = r }
}

func WithStartSessionError(err error) MockOption {
	return func(c *MockClient) { c.StartSessionError = err }
}

func WithUpdateContextError(err error) MockOption {
	return func(c *MockClient) { c.UpdateContextError = err }
}

func WithContextWindowError(err error) MockOption {
	return func(c *MockClient) { c.ContextWindowError = err }
}

func WithEndSessionError(err error) MockOption {
	return func(c *MockClient) { c.EndSessionError = err }
}

func WithRetrieveError(err error) MockOption {
	return func(c *MockClient) { c.RetrieveError = err }
}

func WithRecordEpisodeError(err error) MockOption {
	return func(c *MockClient) { c.RecordEpisodeError = err }
}

func WithRecordFeedbackError(err error) MockOption {
	return func(c *MockClient) { c.RecordFeedbackError = err }
}

func WithStatsError(err error) MockOption {
	return func(c *MockClient) { c.StatsError = err }
}

func WithConsolidationError(err error) MockOption {
	return func(c *MockClient) { c.ConsolidationError = err }
}

func WithHealthError(err error) MockOption {
	return func(c *MockClient) { c.HealthError = err }
}

// Compile-time interface check.
var _ hivemind.Client = (*MockClient)(nil)
