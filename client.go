package hivemind

import (
	"context"
	"fmt"
)

type Client interface {
	StartSession(ctx context.Context, req *StartSessionRequest) (*StartSessionResponse, error)
	UpdateContext(ctx context.Context, req *UpdateContextRequest) error
	GetContextWindow(ctx context.Context, req *ContextWindowRequest) (*ContextWindowResponse, error)
	EndSession(ctx context.Context, runID string) (*EndSessionResponse, error)
	Retrieve(ctx context.Context, req *RetrieveRequest) (*RetrieveResponse, error)
	RecordEpisode(ctx context.Context, req *RecordEpisodeRequest) (*RecordEpisodeResponse, error)
	RecordFeedback(ctx context.Context, req *RecordFeedbackRequest) error
	GetStats(ctx context.Context) (*StatsResponse, error)
	TriggerConsolidation(ctx context.Context) (*ConsolidationResponse, error)
	HealthCheck(ctx context.Context) (*HealthResponse, error)
	Close() error
}

type transport interface {
	StartSession(ctx context.Context, req *StartSessionRequest) (*StartSessionResponse, error)
	UpdateContext(ctx context.Context, req *UpdateContextRequest) error
	GetContextWindow(ctx context.Context, req *ContextWindowRequest) (*ContextWindowResponse, error)
	EndSession(ctx context.Context, runID string) (*EndSessionResponse, error)
	Retrieve(ctx context.Context, req *RetrieveRequest) (*RetrieveResponse, error)
	RecordEpisode(ctx context.Context, req *RecordEpisodeRequest) (*RecordEpisodeResponse, error)
	RecordFeedback(ctx context.Context, req *RecordFeedbackRequest) error
	GetStats(ctx context.Context) (*StatsResponse, error)
	TriggerConsolidation(ctx context.Context) (*ConsolidationResponse, error)
	HealthCheck(ctx context.Context) (*HealthResponse, error)
	Close() error
}

type clientImpl struct {
	config clientConfig
	tp     transport
}

func NewClient(apiKey string, opts ...Option) (Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey is required")
	}
	cfg := defaultConfig(apiKey)
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.endpoint == "" {
		cfg.endpoint = detectEndpoint(apiKey, cfg.transport)
	}

	var tp transport
	var err error
	switch cfg.transport {
	case TransportGRPC:
		tp, err = newGRPCTransport(cfg)
	case TransportREST:
		tp, err = newRESTTransport(cfg)
	default:
		return nil, fmt.Errorf("unsupported transport: %d", cfg.transport)
	}
	if err != nil {
		return nil, fmt.Errorf("creating transport: %w", err)
	}
	return &clientImpl{config: cfg, tp: tp}, nil
}

func (c *clientImpl) StartSession(ctx context.Context, req *StartSessionRequest) (*StartSessionResponse, error) {
	return c.tp.StartSession(ctx, req)
}

func (c *clientImpl) UpdateContext(ctx context.Context, req *UpdateContextRequest) error {
	return c.tp.UpdateContext(ctx, req)
}

func (c *clientImpl) GetContextWindow(ctx context.Context, req *ContextWindowRequest) (*ContextWindowResponse, error) {
	return c.tp.GetContextWindow(ctx, req)
}

func (c *clientImpl) EndSession(ctx context.Context, runID string) (*EndSessionResponse, error) {
	return c.tp.EndSession(ctx, runID)
}

func (c *clientImpl) Retrieve(ctx context.Context, req *RetrieveRequest) (*RetrieveResponse, error) {
	return c.tp.Retrieve(ctx, req)
}

func (c *clientImpl) RecordEpisode(ctx context.Context, req *RecordEpisodeRequest) (*RecordEpisodeResponse, error) {
	return c.tp.RecordEpisode(ctx, req)
}

func (c *clientImpl) RecordFeedback(ctx context.Context, req *RecordFeedbackRequest) error {
	return c.tp.RecordFeedback(ctx, req)
}

func (c *clientImpl) GetStats(ctx context.Context) (*StatsResponse, error) {
	return c.tp.GetStats(ctx)
}

func (c *clientImpl) TriggerConsolidation(ctx context.Context) (*ConsolidationResponse, error) {
	return c.tp.TriggerConsolidation(ctx)
}

func (c *clientImpl) HealthCheck(ctx context.Context) (*HealthResponse, error) {
	return c.tp.HealthCheck(ctx)
}

func (c *clientImpl) Close() error {
	return c.tp.Close()
}
