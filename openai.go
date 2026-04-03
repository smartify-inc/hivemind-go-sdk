//go:build openai

package hivemind

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// WrappedOpenAI wraps an [openai.Client] with automatic Hivemind context
// injection and response recording.
type WrappedOpenAI struct {
	Inner   openai.Client
	Session *AgentSession
}

// WrapOpenAI wraps an OpenAI client with Hivemind hooks.
func WrapOpenAI(openaiClient openai.Client, hivemindClient Client, opts ...SessionOption) *WrappedOpenAI {
	return &WrappedOpenAI{
		Inner:   openaiClient,
		Session: NewSession(hivemindClient, opts...),
	}
}

// CreateChatCompletion calls the OpenAI Chat Completions API with Hivemind
// context prepended as a system message, then records the response.
func (w *WrappedOpenAI) CreateChatCompletion(
	ctx context.Context,
	params openai.ChatCompletionNewParams,
	opts ...option.RequestOption,
) (*openai.ChatCompletion, error) {
	hmContext, err := w.Session.GetContext(ctx, 4000)
	if err != nil {
		return nil, fmt.Errorf("hivemind get context: %w", err)
	}

	if hmContext != "" {
		params.Messages = append(
			[]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(hmContext)},
			params.Messages...,
		)
	}

	resp, err := w.Inner.Chat.Completions.New(ctx, params, opts...)
	if err != nil {
		return resp, err
	}

	if len(resp.Choices) > 0 {
		if recordErr := w.Session.RecordResponse(
			ctx, resp.Choices[0].Message.Content, int(resp.Usage.TotalTokens),
		); recordErr != nil {
			return resp, fmt.Errorf("hivemind record response: %w", recordErr)
		}
	}

	return resp, nil
}

// End ends the Hivemind session and records the episode.
func (w *WrappedOpenAI) End(ctx context.Context) (*EndSessionResponse, error) {
	return w.Session.End(ctx)
}
