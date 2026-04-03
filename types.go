package hivemind

import "time"

type PlanStep struct {
	Index       int    `json:"index"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

type Plan struct {
	Steps        []PlanStep `json:"steps"`
	CurrentIndex int        `json:"current_index"`
}

type StartSessionRequest struct {
	RunID      string `json:"run_id"`
	WorkflowID string `json:"workflow_id"`
	Task       string `json:"task"`
	Plan       *Plan  `json:"plan,omitempty"`
}

type StartSessionResponse struct {
	RunID      string    `json:"run_id"`
	HivemindID string    `json:"hivemind_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type UpdateContextRequest struct {
	RunID string `json:"run_id"`
	Field string `json:"field"`
	Data  any    `json:"data"`
}

type ContextMode string

const (
	ContextModeAuto  ContextMode = "auto"
	ContextModeFull  ContextMode = "full"
	ContextModeDelta ContextMode = "delta"
)

type ContextWindowRequest struct {
	RunID          string      `json:"run_id"`
	AgentID        string      `json:"agent_id"`
	MaxTokens      int         `json:"max_tokens"`
	ConversationID string      `json:"conversation_id,omitempty"`
	ContextMode    ContextMode `json:"context_mode,omitempty"`
}

type ContextWindowResponse struct {
	Context           string `json:"context"`
	ModeUsed          string `json:"mode_used"`
	TurnNumber        int    `json:"turn_number"`
	TotalTokens       int    `json:"total_tokens"`
	DeltaTokensSaved  int    `json:"delta_tokens_saved"`
	FullContextTokens int    `json:"full_context_tokens"`
}

type TokensSaved struct {
	DeltaContext   int64 `json:"delta_context"`
	Compression    int64 `json:"compression"`
	CrossExecution int64 `json:"cross_execution"`
}

type EndSessionResponse struct {
	RunID            string       `json:"run_id"`
	Status           string       `json:"status"`
	EpisodeID        string       `json:"episode_id,omitempty"`
	RecordingStatus  string       `json:"recording_status,omitempty"`
	CompressionRatio float64      `json:"compression_ratio,omitempty"`
	TokensSaved      TokensSaved  `json:"tokens_saved"`
	Warning          string       `json:"warning,omitempty"`
}

type RetrieveRequest struct {
	Intent         string   `json:"intent"`
	TaskType       string   `json:"task_type,omitempty"`
	Constraints    []string `json:"constraints,omitempty"`
	MaxTokens      int      `json:"max_tokens,omitempty"`
	Depth          int      `json:"depth,omitempty"`
	MinConfidence  float64  `json:"min_confidence,omitempty"`
	ScopeType      string   `json:"scope_type,omitempty"`
	ScopeID        string   `json:"scope_id,omitempty"`
	IncludeUnscoped *bool   `json:"include_unscoped,omitempty"`
}

type RetrieveResponse struct {
	ContextText        string  `json:"context_text"`
	ContextTokens      int     `json:"context_tokens"`
	BudgetUsed         float64 `json:"budget_used"`
	SkillsCount        int     `json:"skills_count"`
	PatternsCount      int     `json:"patterns_count"`
	EpisodesCount      int     `json:"episodes_count"`
	RetrievalLatencyMs float64 `json:"retrieval_latency_ms"`
	TokensSavedEstimate int64  `json:"tokens_saved_estimate"`
}

type StepResult struct {
	Action     string `json:"action"`
	Outcome    string `json:"outcome"`
	TokensUsed int    `json:"tokens_used"`
	DurationMs int    `json:"duration_ms"`
	Rationale  string `json:"rationale,omitempty"`
}

type RecordEpisodeRequest struct {
	WorkflowID      string                `json:"workflow_id"`
	RunID           string                `json:"run_id"`
	TaskDescription string                `json:"task_description"`
	StepResults     map[string]StepResult `json:"step_results"`
	Inputs          map[string]any        `json:"inputs,omitempty"`
	Outputs         map[string]any        `json:"outputs,omitempty"`
	TotalTokens     int                   `json:"total_tokens"`
	TotalCost       float64               `json:"total_cost"`
	DurationMs      int                   `json:"duration_ms"`
	ScopeType       string                `json:"scope_type,omitempty"`
	ScopeID         string                `json:"scope_id,omitempty"`
}

type RecordEpisodeResponse struct {
	EpisodeID         string      `json:"episode_id"`
	TaskType          string      `json:"task_type"`
	CompressionRatio  float64     `json:"compression_ratio"`
	InsightsExtracted int         `json:"insights_extracted"`
	TokensSaved       TokensSaved `json:"tokens_saved"`
}

type RecordFeedbackRequest struct {
	EpisodeID string  `json:"episode_id"`
	Score     float64 `json:"score"`
	Notes     string  `json:"notes,omitempty"`
}

type TokensSavedBySource struct {
	DeltaContext   int64 `json:"delta_context"`
	CrossExecution int64 `json:"cross_execution"`
	Compression    int64 `json:"compression"`
}

type TokensSavedStats struct {
	Total          int64               `json:"total"`
	BySource       TokensSavedBySource `json:"by_source"`
	Period         string              `json:"period"`
	CostSavingsUSD float64             `json:"cost_savings_usd"`
}

type KnowledgeStateDistribution struct {
	Embryonic int `json:"embryonic"`
	Growing   int `json:"growing"`
	Mature    int `json:"mature"`
}

type StatsResponse struct {
	TotalEpisodes              int                        `json:"total_episodes"`
	TotalKnowledgeNodes        int                        `json:"total_knowledge_nodes"`
	TotalEdges                 int                        `json:"total_edges"`
	CompressionRatioAvg        float64                    `json:"compression_ratio_avg"`
	TokensSaved                TokensSavedStats            `json:"tokens_saved"`
	ActiveSessions             int                        `json:"active_sessions"`
	LastConsolidationAt        *time.Time                 `json:"last_consolidation_at,omitempty"`
	KnowledgeStateDistribution KnowledgeStateDistribution `json:"knowledge_state_distribution"`
}

type ConsolidationResponse struct {
	ConsolidationID string `json:"consolidation_id"`
	Status          string `json:"status"`
	EpisodesQueued  int    `json:"episodes_queued"`
}

type ServiceHealth struct {
	Aurora  string `json:"aurora"`
	Redis   string `json:"redis"`
	Neo4j   string `json:"neo4j"`
	PgVector string `json:"pgvector"`
}

type HealthResponse struct {
	Status   string        `json:"status"`
	Version  string        `json:"version"`
	Services ServiceHealth `json:"services"`
}
