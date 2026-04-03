# Hivemind Go SDK

Go client for [Smartify Hivemind](https://smartify.ai) — persistent memory and context for multi-agent AI. The package speaks to the public Hivemind API over **gRPC** (default) or **REST**.

**Module:** `github.com/smartifyai/hivemind-go`

## Requirements

- Go 1.24+

## Install

```bash
go get github.com/smartifyai/hivemind-go@latest
```

Set your API key (for example `HIVEMIND_API_KEY`). Keys starting with `sk_live_` target production; other keys (e.g. `sk_test_`) use staging endpoints unless you override the endpoint.

## Quick start — `AgentSession` (hooks)

`AgentSession` starts a run, fetches packed context for each LLM turn, records assistant output, and on `End` records the episode.

```go
package main

import (
	"context"
	"log"
	"os"

	hivemind "github.com/smartifyai/hivemind-go"
)

func main() {
	client, err := hivemind.NewClient(os.Getenv("HIVEMIND_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	session := hivemind.NewSession(client,
		hivemind.WithWorkflowID("blog-writer"),
		hivemind.WithTask("Write about caching"),
	)

	contextText, err := session.GetContext(ctx, 4000)
	if err != nil {
		log.Fatal(err)
	}
	_ = contextText // inject into your LLM system / messages

	// After your LLM returns:
	if err := session.RecordResponse(ctx, "assistant reply text here", 150); err != nil {
		log.Fatal(err)
	}

	endResp, err := session.End(ctx)
	if err != nil {
		log.Fatal(err)
	}
	_ = endResp // episode + savings stats
}
```

For conversation-scoped delta context, use `GetContextWithConversation(ctx, maxTokens, conversationID)`.

## Quick start — low-level `Client`

Use `hivemind.NewClient` and the `Client` interface when you want full control over session and context calls. See [specification/GOLANG_SDK.md](../../specification/GOLANG_SDK.md) in this monorepo for a full walkthrough.

## Transports and endpoints

| Option | Default |
|--------|---------|
| Transport | gRPC (`TransportGRPC`) |
| Live API key (`sk_live_…`) | `grpc.smartify.ai:443` or `https://api.smartify.ai` (REST) |
| Non-live key | `grpc-staging.smartify.ai:443` or `https://api-staging.smartify.ai` (REST) |

Switch to REST:

```go
client, err := hivemind.NewClient(apiKey, hivemind.WithTransport(hivemind.TransportREST))
```

Override host explicitly:

```go
client, err := hivemind.NewClient(apiKey,
	hivemind.WithTransport(hivemind.TransportREST),
	hivemind.WithEndpoint("https://api.example.com"),
)
```

Other useful options: `WithTimeout`, `WithHivemindID`, `WithRetryPolicy`, `WithUserAgent`, `WithLogger`.

## OpenAI helper (optional build tag)

Optional integration with [`github.com/sashabaranov/go-openai`](https://github.com/sashabaranov/go-openai) lives behind the `openai` build tag so the dependency stays optional.

```bash
go test -tags=openai ./...
```

```go
import (
	"context"
	openai "github.com/sashabaranov/go-openai"
	hivemind "github.com/smartifyai/hivemind-go"
)

wrapped := hivemind.WrapOpenAI(openaiClient, hivemindClient,
	hivemind.WithWorkflowID("blog-writer"),
	hivemind.WithTask("Write about caching"),
)
resp, err := wrapped.CreateChatCompletion(ctx, openai.ChatCompletionRequest{ /* model, messages, ... */ })
```

Call `wrapped.End(ctx)` when finished to end the Hivemind session.

## Testing

The subpackage `github.com/smartifyai/hivemind-go/mock` provides a `mock.Client` that implements `hivemind.Client` and records calls for tests.

## License

Copyright 2026 Smartify Inc. Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).
