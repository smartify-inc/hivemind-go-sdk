package hivemind

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/smartifyai/hivemind-go/proto/hivemindv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type grpcTransport struct {
	conn   *grpc.ClientConn
	client pb.HivemindServiceClient
	config clientConfig
}

func newGRPCTransport(cfg clientConfig) (*grpcTransport, error) {
	conn, err := grpc.NewClient(
		cfg.endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInterceptor(cfg)),
	)
	if err != nil {
		return nil, fmt.Errorf("dialing gRPC endpoint %s: %w", cfg.endpoint, err)
	}
	return &grpcTransport{
		conn:   conn,
		client: pb.NewHivemindServiceClient(conn),
		config: cfg,
	}, nil
}

func authInterceptor(cfg clientConfig) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		md := metadata.Pairs(
			"authorization", "Bearer "+cfg.apiKey,
		)
		if cfg.hivemindID != "" {
			md.Append("x-hivemind-id", cfg.hivemindID)
		}
		if cfg.userAgent != "" {
			md.Append("user-agent", cfg.userAgent)
		}
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (t *grpcTransport) StartSession(ctx context.Context, req *StartSessionRequest) (*StartSessionResponse, error) {
	pbReq := &pb.StartSessionRequest{
		RunId:      req.RunID,
		WorkflowId: req.WorkflowID,
		Task:       req.Task,
		HivemindId: t.config.hivemindID,
	}
	if req.Plan != nil {
		pbReq.Plan = planToProto(req.Plan)
	}

	resp, err := t.client.StartSession(ctx, pbReq)
	if err != nil {
		return nil, grpcErr(err)
	}

	out := &StartSessionResponse{
		RunID:      resp.GetRunId(),
		HivemindID: resp.GetHivemindId(),
		Status:     resp.GetStatus(),
	}
	if resp.GetCreatedAt() != nil {
		out.CreatedAt = resp.GetCreatedAt().AsTime()
	}
	return out, nil
}

func (t *grpcTransport) UpdateContext(ctx context.Context, req *UpdateContextRequest) error {
	pbReq := &pb.UpdateContextRequest{
		RunId: req.RunID,
		Field: req.Field,
	}
	if req.Data != nil {
		s, err := anyToProtoStruct(req.Data)
		if err != nil {
			return fmt.Errorf("marshalling context data: %w", err)
		}
		pbReq.Data = s
	}

	_, err := t.client.UpdateContext(ctx, pbReq)
	if err != nil {
		return grpcErr(err)
	}
	return nil
}

func (t *grpcTransport) GetContextWindow(ctx context.Context, req *ContextWindowRequest) (*ContextWindowResponse, error) {
	pbReq := &pb.GetContextWindowRequest{
		RunId:          req.RunID,
		AgentId:        req.AgentID,
		MaxTokens:      int32(req.MaxTokens),
		ConversationId: req.ConversationID,
		ContextMode:    string(req.ContextMode),
	}

	resp, err := t.client.GetContextWindow(ctx, pbReq)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &ContextWindowResponse{
		Context:           resp.GetContext(),
		ModeUsed:          resp.GetModeUsed(),
		TurnNumber:        int(resp.GetTurnNumber()),
		TotalTokens:       int(resp.GetTotalTokens()),
		DeltaTokensSaved:  int(resp.GetDeltaTokensSaved()),
		FullContextTokens: int(resp.GetFullContextTokens()),
	}, nil
}

func (t *grpcTransport) EndSession(ctx context.Context, runID string) (*EndSessionResponse, error) {
	resp, err := t.client.EndSession(ctx, &pb.EndSessionRequest{RunId: runID})
	if err != nil {
		return nil, grpcErr(err)
	}

	out := &EndSessionResponse{
		RunID:            resp.GetRunId(),
		Status:           resp.GetStatus(),
		EpisodeID:        resp.GetEpisodeId(),
		RecordingStatus:  resp.GetRecordingStatus(),
		CompressionRatio: resp.GetCompressionRatio(),
		Warning:          resp.GetWarning(),
	}
	if ts := resp.GetTokensSaved(); ts != nil {
		out.TokensSaved = TokensSaved{
			DeltaContext:   ts.GetDeltaContext(),
			Compression:    ts.GetCompression(),
			CrossExecution: ts.GetCrossExecution(),
		}
	}
	return out, nil
}

func (t *grpcTransport) Retrieve(ctx context.Context, req *RetrieveRequest) (*RetrieveResponse, error) {
	pbReq := &pb.RetrieveRequest{
		Intent:        req.Intent,
		TaskType:      req.TaskType,
		Constraints:   req.Constraints,
		MaxTokens:     int32(req.MaxTokens),
		Depth:         int32(req.Depth),
		MinConfidence: req.MinConfidence,
		ScopeType:     req.ScopeType,
		ScopeId:       req.ScopeID,
	}
	if req.IncludeUnscoped != nil {
		pbReq.IncludeUnscoped = *req.IncludeUnscoped
	}

	resp, err := t.client.Retrieve(ctx, pbReq)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &RetrieveResponse{
		ContextText:        resp.GetContextText(),
		ContextTokens:      int(resp.GetContextTokens()),
		BudgetUsed:         resp.GetBudgetUsed(),
		SkillsCount:        int(resp.GetSkillsCount()),
		PatternsCount:      int(resp.GetPatternsCount()),
		EpisodesCount:      int(resp.GetEpisodesCount()),
		RetrievalLatencyMs: resp.GetRetrievalLatencyMs(),
		TokensSavedEstimate: resp.GetTokensSavedEstimate(),
	}, nil
}

func (t *grpcTransport) RecordEpisode(ctx context.Context, req *RecordEpisodeRequest) (*RecordEpisodeResponse, error) {
	pbReq := &pb.RecordEpisodeRequest{
		WorkflowId:      req.WorkflowID,
		RunId:           req.RunID,
		TaskDescription: req.TaskDescription,
		TotalTokens:     int32(req.TotalTokens),
		TotalCost:       req.TotalCost,
		DurationMs:      int32(req.DurationMs),
		ScopeType:       req.ScopeType,
		ScopeId:         req.ScopeID,
	}
	if req.StepResults != nil {
		pbReq.StepResults = make(map[string]*pb.StepResult, len(req.StepResults))
		for k, v := range req.StepResults {
			pbReq.StepResults[k] = &pb.StepResult{
				Action:     v.Action,
				Outcome:    v.Outcome,
				TokensUsed: int32(v.TokensUsed),
				DurationMs: int32(v.DurationMs),
				Rationale:  v.Rationale,
			}
		}
	}
	if req.Inputs != nil {
		s, err := structpb.NewStruct(req.Inputs)
		if err != nil {
			return nil, fmt.Errorf("marshalling inputs: %w", err)
		}
		pbReq.Inputs = s
	}
	if req.Outputs != nil {
		s, err := structpb.NewStruct(req.Outputs)
		if err != nil {
			return nil, fmt.Errorf("marshalling outputs: %w", err)
		}
		pbReq.Outputs = s
	}

	resp, err := t.client.RecordEpisode(ctx, pbReq)
	if err != nil {
		return nil, grpcErr(err)
	}

	out := &RecordEpisodeResponse{
		EpisodeID:         resp.GetEpisodeId(),
		TaskType:          resp.GetTaskType(),
		CompressionRatio:  resp.GetCompressionRatio(),
		InsightsExtracted: int(resp.GetInsightsExtracted()),
	}
	if ts := resp.GetTokensSaved(); ts != nil {
		out.TokensSaved = TokensSaved{
			DeltaContext:   ts.GetDeltaContext(),
			Compression:    ts.GetCompression(),
			CrossExecution: ts.GetCrossExecution(),
		}
	}
	return out, nil
}

func (t *grpcTransport) RecordFeedback(ctx context.Context, req *RecordFeedbackRequest) error {
	_, err := t.client.RecordFeedback(ctx, &pb.RecordFeedbackRequest{
		EpisodeId: req.EpisodeID,
		Score:     req.Score,
		Notes:     req.Notes,
	})
	if err != nil {
		return grpcErr(err)
	}
	return nil
}

func (t *grpcTransport) GetStats(ctx context.Context) (*StatsResponse, error) {
	resp, err := t.client.GetStats(ctx, &pb.GetStatsRequest{})
	if err != nil {
		return nil, grpcErr(err)
	}

	out := &StatsResponse{
		TotalEpisodes:       int(resp.GetTotalEpisodes()),
		TotalKnowledgeNodes: int(resp.GetTotalKnowledgeNodes()),
		TotalEdges:          int(resp.GetTotalEdges()),
		CompressionRatioAvg: resp.GetCompressionRatioAvg(),
		ActiveSessions:      int(resp.GetActiveSessions()),
	}
	if ts := resp.GetTokensSaved(); ts != nil {
		out.TokensSaved = tokensSavedStatsFromProto(ts)
	}
	if resp.GetLastConsolidationAt() != nil {
		t := resp.GetLastConsolidationAt().AsTime()
		out.LastConsolidationAt = &t
	}
	if ksd := resp.GetKnowledgeStateDistribution(); ksd != nil {
		out.KnowledgeStateDistribution = KnowledgeStateDistribution{
			Embryonic: int(ksd.GetEmbryonic()),
			Growing:   int(ksd.GetGrowing()),
			Mature:    int(ksd.GetMature()),
		}
	}
	return out, nil
}

func (t *grpcTransport) TriggerConsolidation(ctx context.Context) (*ConsolidationResponse, error) {
	resp, err := t.client.TriggerConsolidation(ctx, &pb.TriggerConsolidationRequest{})
	if err != nil {
		return nil, grpcErr(err)
	}
	return &ConsolidationResponse{
		ConsolidationID: resp.GetConsolidationId(),
		Status:          resp.GetStatus(),
		EpisodesQueued:  int(resp.GetEpisodesQueued()),
	}, nil
}

func (t *grpcTransport) HealthCheck(ctx context.Context) (*HealthResponse, error) {
	resp, err := t.client.HealthCheck(ctx, &pb.HealthCheckRequest{})
	if err != nil {
		return nil, grpcErr(err)
	}

	out := &HealthResponse{
		Status:  resp.GetStatus(),
		Version: resp.GetVersion(),
	}
	if svc := resp.GetServices(); svc != nil {
		out.Services = ServiceHealth{
			Aurora:   svc.GetAurora(),
			Redis:    svc.GetRedis(),
			Neo4j:    svc.GetNeo4J(),
			PgVector: svc.GetPgvector(),
		}
	}
	return out, nil
}

func (t *grpcTransport) Close() error {
	return t.conn.Close()
}

// --- helpers ---

func planToProto(p *Plan) *pb.Plan {
	out := &pb.Plan{CurrentIndex: int32(p.CurrentIndex)}
	for _, s := range p.Steps {
		out.Steps = append(out.Steps, &pb.PlanStep{
			Index:       int32(s.Index),
			Action:      s.Action,
			Description: s.Description,
		})
	}
	return out
}

func anyToProtoStruct(v any) (*structpb.Struct, error) {
	switch d := v.(type) {
	case map[string]any:
		return structpb.NewStruct(d)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("json marshal: %w", err)
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			m = map[string]any{"value": v}
			return structpb.NewStruct(m)
		}
		return structpb.NewStruct(m)
	}
}

func tokensSavedStatsFromProto(ts *pb.TokensSavedStats) TokensSavedStats {
	out := TokensSavedStats{
		Total:          ts.GetTotal(),
		Period:         ts.GetPeriod(),
		CostSavingsUSD: ts.GetCostSavingsUsd(),
	}
	if bs := ts.GetBySource(); bs != nil {
		out.BySource = TokensSavedBySource{
			DeltaContext:   bs.GetDeltaContext(),
			CrossExecution: bs.GetCrossExecution(),
			Compression:    bs.GetCompression(),
		}
	}
	return out
}

var grpcCodeToHTTP = map[codes.Code]int{
	codes.OK:                 200,
	codes.Canceled:           499,
	codes.InvalidArgument:    400,
	codes.NotFound:           404,
	codes.AlreadyExists:      409,
	codes.PermissionDenied:   403,
	codes.Unauthenticated:    401,
	codes.ResourceExhausted:  429,
	codes.FailedPrecondition: 412,
	codes.Aborted:            409,
	codes.OutOfRange:         400,
	codes.Unimplemented:      501,
	codes.Internal:           500,
	codes.Unavailable:        503,
	codes.DataLoss:           500,
	codes.DeadlineExceeded:   504,
	codes.Unknown:            500,
}

func grpcErr(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	httpCode, mapped := grpcCodeToHTTP[st.Code()]
	if !mapped {
		httpCode = 500
	}
	return &HivemindError{
		Type:   "grpc_error",
		Title:  st.Code().String(),
		Status: httpCode,
		Detail: st.Message(),
	}
}

// compile-time interface check
var _ transport = (*grpcTransport)(nil)

