package hivemind

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type restTransport struct {
	httpClient *http.Client
	baseURL    string
	config     clientConfig
}

func newRESTTransport(cfg clientConfig) (*restTransport, error) {
	base := strings.TrimRight(cfg.endpoint, "/")
	return &restTransport{
		httpClient: &http.Client{Timeout: cfg.timeout},
		baseURL:    base,
		config:     cfg,
	}, nil
}

func (t *restTransport) StartSession(ctx context.Context, req *StartSessionRequest) (*StartSessionResponse, error) {
	var resp StartSessionResponse
	if err := t.do(ctx, http.MethodPost, "/v1/sessions", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t *restTransport) UpdateContext(ctx context.Context, req *UpdateContextRequest) error {
	path := fmt.Sprintf("/v1/sessions/%s/context", url.PathEscape(req.RunID))
	return t.do(ctx, http.MethodPut, path, req, nil)
}

func (t *restTransport) GetContextWindow(ctx context.Context, req *ContextWindowRequest) (*ContextWindowResponse, error) {
	path := fmt.Sprintf("/v1/sessions/%s/context-window", url.PathEscape(req.RunID))

	q := url.Values{}
	q.Set("agent_id", req.AgentID)
	if req.MaxTokens > 0 {
		q.Set("max_tokens", fmt.Sprintf("%d", req.MaxTokens))
	}
	if req.ConversationID != "" {
		q.Set("conversation_id", req.ConversationID)
	}
	if req.ContextMode != "" {
		q.Set("context_mode", string(req.ContextMode))
	}
	path += "?" + q.Encode()

	var resp ContextWindowResponse
	if err := t.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t *restTransport) EndSession(ctx context.Context, runID string) (*EndSessionResponse, error) {
	path := fmt.Sprintf("/v1/sessions/%s", url.PathEscape(runID))
	var resp EndSessionResponse
	if err := t.do(ctx, http.MethodDelete, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t *restTransport) Retrieve(ctx context.Context, req *RetrieveRequest) (*RetrieveResponse, error) {
	var resp RetrieveResponse
	if err := t.do(ctx, http.MethodPost, "/v1/retrieve", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t *restTransport) RecordEpisode(ctx context.Context, req *RecordEpisodeRequest) (*RecordEpisodeResponse, error) {
	var resp RecordEpisodeResponse
	if err := t.do(ctx, http.MethodPost, "/v1/episodes", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t *restTransport) RecordFeedback(ctx context.Context, req *RecordFeedbackRequest) error {
	path := fmt.Sprintf("/v1/episodes/%s/feedback", url.PathEscape(req.EpisodeID))
	return t.do(ctx, http.MethodPost, path, req, nil)
}

func (t *restTransport) GetStats(ctx context.Context) (*StatsResponse, error) {
	var resp StatsResponse
	if err := t.do(ctx, http.MethodGet, "/v1/stats", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t *restTransport) TriggerConsolidation(ctx context.Context) (*ConsolidationResponse, error) {
	var resp ConsolidationResponse
	if err := t.do(ctx, http.MethodPost, "/v1/consolidation/trigger", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t *restTransport) HealthCheck(ctx context.Context) (*HealthResponse, error) {
	var resp HealthResponse
	if err := t.do(ctx, http.MethodGet, "/v1/health", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t *restTransport) Close() error {
	return nil
}

// do executes an HTTP request with auth headers and JSON encoding.
func (t *restTransport) do(ctx context.Context, method, path string, body, result any) error {
	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshalling request body: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	reqURL := t.baseURL + path
	httpReq, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+t.config.apiKey)
	if t.config.hivemindID != "" {
		httpReq.Header.Set("X-Hivemind-ID", t.config.hivemindID)
	}
	if t.config.userAgent != "" {
		httpReq.Header.Set("User-Agent", t.config.userAgent)
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("executing HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseHTTPError(resp.StatusCode, respBody)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshalling response: %w", err)
		}
	}
	return nil
}

func parseHTTPError(statusCode int, body []byte) error {
	var he HivemindError
	if err := json.Unmarshal(body, &he); err == nil && he.Title != "" {
		he.Status = statusCode
		return &he
	}
	return &HivemindError{
		Type:   "http_error",
		Title:  http.StatusText(statusCode),
		Status: statusCode,
		Detail: string(body),
	}
}

var _ transport = (*restTransport)(nil)
