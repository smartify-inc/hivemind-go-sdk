package hivemind

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// AgentSession manages a Hivemind session for a single agent execution run.
// It tracks context retrieval and response recording across multiple LLM turns,
// then records the full episode when End() is called.
type AgentSession struct {
	client     Client
	workflowID string
	task       string
	agentID    string
	runID      string

	mu          sync.Mutex
	started     bool
	turn        int
	totalTokens int
	stepResults map[string]StepResult
}

// SessionOption configures an AgentSession.
type SessionOption func(*AgentSession)

func WithWorkflowID(id string) SessionOption {
	return func(s *AgentSession) { s.workflowID = id }
}

func WithTask(task string) SessionOption {
	return func(s *AgentSession) { s.task = task }
}

func WithAgentID(id string) SessionOption {
	return func(s *AgentSession) { s.agentID = id }
}

func WithRunID(id string) SessionOption {
	return func(s *AgentSession) { s.runID = id }
}

// NewSession creates a new AgentSession with the given client and options.
func NewSession(client Client, opts ...SessionOption) *AgentSession {
	s := &AgentSession{
		client:      client,
		agentID:     "default",
		stepResults: make(map[string]StepResult),
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.runID == "" {
		s.runID = uuid.New().String()
	}
	return s
}

// RunID returns the session's run ID.
func (s *AgentSession) RunID() string { return s.runID }

// WorkflowID returns the session's workflow ID.
func (s *AgentSession) WorkflowID() string { return s.workflowID }

// Turn returns the current turn number.
func (s *AgentSession) Turn() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.turn
}

func (s *AgentSession) ensureStarted(ctx context.Context) error {
	s.mu.Lock()
	started := s.started
	s.mu.Unlock()

	if started {
		return nil
	}

	_, err := s.client.StartSession(ctx, &StartSessionRequest{
		RunID:      s.runID,
		WorkflowID: s.workflowID,
		Task:       s.task,
	})
	if err != nil {
		return fmt.Errorf("starting session: %w", err)
	}

	s.mu.Lock()
	s.started = true
	s.mu.Unlock()
	return nil
}

// GetContext retrieves packed context for the next LLM call.
// Automatically starts the session on first call.
func (s *AgentSession) GetContext(ctx context.Context, maxTokens int) (string, error) {
	if err := s.ensureStarted(ctx); err != nil {
		return "", err
	}

	resp, err := s.client.GetContextWindow(ctx, &ContextWindowRequest{
		RunID:     s.runID,
		AgentID:   s.agentID,
		MaxTokens: maxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("getting context window: %w", err)
	}
	return resp.Context, nil
}

// GetContextWithConversation retrieves packed context scoped to a conversation.
func (s *AgentSession) GetContextWithConversation(ctx context.Context, maxTokens int, conversationID string) (string, error) {
	if err := s.ensureStarted(ctx); err != nil {
		return "", err
	}

	resp, err := s.client.GetContextWindow(ctx, &ContextWindowRequest{
		RunID:          s.runID,
		AgentID:        s.agentID,
		MaxTokens:      maxTokens,
		ConversationID: conversationID,
	})
	if err != nil {
		return "", fmt.Errorf("getting context window: %w", err)
	}
	return resp.Context, nil
}

// RecordResponse records an LLM response and tracks step results.
func (s *AgentSession) RecordResponse(ctx context.Context, response string, tokensUsed int) error {
	if err := s.ensureStarted(ctx); err != nil {
		return err
	}

	s.mu.Lock()
	s.turn++
	s.totalTokens += tokensUsed
	stepKey := fmt.Sprintf("turn_%d", s.turn)
	outcome := response
	if len(outcome) > 500 {
		outcome = outcome[:500]
	}
	s.stepResults[stepKey] = StepResult{
		Action:     "llm_call",
		Outcome:    outcome,
		TokensUsed: tokensUsed,
	}
	s.mu.Unlock()

	return s.client.UpdateContext(ctx, &UpdateContextRequest{
		RunID: s.runID,
		Field: "messages",
		Data: map[string]string{
			"role":    "assistant",
			"content": response,
		},
	})
}

// AddObservation adds a note/observation to the session context.
func (s *AgentSession) AddObservation(ctx context.Context, observation string) error {
	if err := s.ensureStarted(ctx); err != nil {
		return err
	}

	return s.client.UpdateContext(ctx, &UpdateContextRequest{
		RunID: s.runID,
		Field: "observations",
		Data:  observation,
	})
}

// AddToolResult records a tool call result in the session.
func (s *AgentSession) AddToolResult(ctx context.Context, tool, result string) error {
	if err := s.ensureStarted(ctx); err != nil {
		return err
	}

	s.mu.Lock()
	s.turn++
	stepKey := fmt.Sprintf("tool_%d", s.turn)
	outcome := result
	if len(outcome) > 500 {
		outcome = outcome[:500]
	}
	s.stepResults[stepKey] = StepResult{
		Action:  "tool:" + tool,
		Outcome: outcome,
	}
	s.mu.Unlock()

	return s.client.UpdateContext(ctx, &UpdateContextRequest{
		RunID: s.runID,
		Field: "tool_results",
		Data: map[string]string{
			"tool":   tool,
			"result": result,
		},
	})
}

// End ends the session, records the episode, and returns savings stats.
func (s *AgentSession) End(ctx context.Context) (*EndSessionResponse, error) {
	s.mu.Lock()
	started := s.started
	stepResults := s.stepResults
	totalTokens := s.totalTokens
	s.mu.Unlock()

	if !started {
		return &EndSessionResponse{RunID: s.runID, Status: "not_started"}, nil
	}

	if len(stepResults) > 0 {
		_, err := s.client.RecordEpisode(ctx, &RecordEpisodeRequest{
			WorkflowID:      s.workflowID,
			RunID:           s.runID,
			TaskDescription: s.task,
			StepResults:     stepResults,
			TotalTokens:     totalTokens,
		})
		if err != nil {
			return nil, fmt.Errorf("recording episode: %w", err)
		}
	}

	resp, err := s.client.EndSession(ctx, s.runID)
	if err != nil {
		return nil, fmt.Errorf("ending session: %w", err)
	}

	s.mu.Lock()
	s.started = false
	s.mu.Unlock()

	return resp, nil
}
