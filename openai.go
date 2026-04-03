//go:build openai

package hivemind

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// WrappedOpenAI wraps an OpenAI client with automatic Hivemind context
// injection and response recording.
type WrappedOpenAI struct {
	inner   *openai.Client
	session *AgentSession
}

// WrapOpenAI wraps an OpenAI client with Hivemind hooks.
// The returned WrappedOpenAI will automatically inject context into prompts
// and record responses on every CreateChatCompletion call.
func WrapOpenAI(openaiClient *openai.Client, hivemindClient Client, opts ...SessionOption) *WrappedOpenAI {
	session := NewSession(hivemindClient, opts...)
	return &WrappedOpenAI{
		inner:   openaiClient,
		session: session,
	}
}

// Session returns the underlying AgentSession.
func (w *WrappedOpenAI) Session() *AgentSession {
	return w.session
}

// Inner returns the underlying OpenAI client for direct access.
func (w *WrappedOpenAI) Inner() *openai.Client {
	return w.inner
}

// CreateChatCompletion calls the OpenAI API with Hivemind context auto-injected
// as a system message prepended to the request messages. The response is
// automatically recorded back to Hivemind.
func (w *WrappedOpenAI) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	hmContext, err := w.session.GetContext(ctx, 4000)
	if err != nil {
		return openai.ChatCompletionResponse{}, fmt.Errorf("hivemind get context: %w", err)
	}

	if hmContext != "" {
		contextMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: hmContext,
		}
		req.Messages = append([]openai.ChatCompletionMessage{contextMsg}, req.Messages...)
	}

	resp, err := w.inner.CreateChatCompletion(ctx, req)
	if err != nil {
		return resp, err
	}

	if len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content
		tokensUsed := resp.Usage.TotalTokens
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
