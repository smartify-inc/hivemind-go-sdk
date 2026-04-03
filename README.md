# Hivemind Go SDK

Go client for [Smartify Hivemind](https://smartify.ai) — persistent memory and context for multi-agent AI. The package speaks to the public Hivemind API over **gRPC** (default) or **REST**.

**Module:** `github.com/smartify-inc/hivemind-go-sdk`

## Requirements

- Go 1.26+

## Install

```bash
go get github.com/smartify-inc/hivemind-go-sdk@latest
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

	hivemind "github.com/smartify-inc/hivemind-go-sdk"
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

Use `hivemind.NewClient` and the `Client` interface when you want full control over session and context calls. Reference docs: [pkg.go.dev/github.com/smartify-inc/hivemind-go-sdk](https://pkg.go.dev/github.com/smartify-inc/hivemind-go-sdk) and the tests in this repository for examples.

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

Optional integration with the official OpenAI Go SDK ([`openai/openai-go`](https://github.com/openai/openai-go)) lives behind the `openai` build tag. The module still lists that dependency so `go mod download` resolves it; only the wrapper sources are tag-gated.

```bash
go test -tags=openai ./...
```

```go
import (
	"context"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	hivemind "github.com/smartify-inc/hivemind-go-sdk"
)

hmClient, _ := hivemind.NewClient(os.Getenv("HIVEMIND_API_KEY"))
defer hmClient.Close()

oai := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))

wrapped := hivemind.WrapOpenAI(oai, hmClient,
	hivemind.WithWorkflowID("blog-writer"),
	hivemind.WithTask("Write about caching"),
)

resp, err := wrapped.CreateChatCompletion(ctx, openai.ChatCompletionNewParams{
	Model: openai.ChatModelGPT4o,
	Messages: []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Write a haiku about caching."),
	},
})
```

Call `wrapped.End(ctx)` when finished to end the Hivemind session. The `Inner` and `Session` fields are exported if you need direct access.

## Testing

The subpackage `github.com/smartify-inc/hivemind-go-sdk/mock` provides a `mock.Client` that implements `hivemind.Client` and records calls for tests.

## License

Copyright 2026 Smartify Inc. Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).
