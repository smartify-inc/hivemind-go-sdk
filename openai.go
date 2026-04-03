//go:build openai

package hivemind

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// WrappedOpenAI wraps the official OpenAI Go client ([openai.Client]) with
// automatic Hivemind context injection and response recording.
//
// See: https://github.com/openai/openai-go
type WrappedOpenAI struct {
	inner         openai.Client
	session       *AgentSession
	contextMaxTok int
}

// WrapOpenAI wraps an OpenAI client from [github.com/openai/openai-go/v3] with Hivemind hooks.
// Pass the value returned by [openai.NewClient] (or a copy with your options).
//
// contextMaxTokens sets the max token budget passed to [AgentSession.GetContext] on each
// completion (default 4000 if zero).
func WrapOpenAI(openaiClient openai.Client, hivemindClient Client, opts ...SessionOption) *WrappedOpenAI {
	return WrapOpenAIWithContextLimit(openaiClient, hivemindClient, 4000, opts...)
}

// WrapOpenAIWithContextLimit is like [WrapOpenAI] but allows setting the Hivemind context window size.
func WrapOpenAIWithContextLimit(openaiClient openai.Client, hivemindClient Client, contextMaxTokens int, opts ...SessionOption) *WrappedOpenAI {
	maxTok := contextMaxTokens
	if maxTok <= 0 {
		maxTok = 4000
	}
	session := NewSession(hivemindClient, opts...)
	return &WrappedOpenAI{
		inner:         openaiClient,
		session:       session,
		contextMaxTok: maxTok,
	}
}

// Session returns the underlying AgentSession.
func (w *WrappedOpenAI) Session() *AgentSession {
	return w.session
}

// Inner returns the wrapped OpenAI client for direct API access.
func (w *WrappedOpenAI) Inner() openai.Client {
	return w.inner
}

// CreateChatCompletion calls [openai.ChatCompletionService.New] with Hivemind context prepended
// as a system message, then records the assistant reply via [AgentSession.RecordResponse].
//
// Extra [option.RequestOption] values are forwarded to the OpenAI client (retries, headers, etc.).
func (w *WrappedOpenAI) CreateChatCompletion(
	ctx context.Context,
	params openai.ChatCompletionNewParams,
	opts ...option.RequestOption,
) (*openai.ChatCompletion, error) {
	hmContext, err := w.session.GetContext(ctx, w.contextMaxTok)
	if err != nil {
		return nil, fmt.Errorf("hivemind get context: %w", err)
	}

	if hmContext != "" {
		systemMsg := openai.SystemMessage(hmContext)
		params.Messages = append([]openai.ChatCompletionMessageParamUnion{systemMsg}, params.Messages...)
	}

	resp, err := w.inner.Chat.Completions.New(ctx, params, opts...)
	if err != nil {
		return resp, err
	}

	if len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content
		tokensUsed := int(resp.Usage.TotalTokens)
		if recordErr := w.session.RecordResponse(ctx, content, tokensUsed); recordErr != nil {
			return resp, fmt.Errorf("hivemind record response: %w", recordErr)
		}
	}

	return resp, nil
}

// End ends the Hivemind session and records the episode.
func (w *WrappedOpenAI) End(ctx context.Context) (*EndSessionResponse, error) {
	return w.session.End(ctx)
}
