# Hivemind Go SDK

Go client for [Smartify Hivemind](https://smartify.ai) — memory and context for AI agents. Talks to the Hivemind API over gRPC (default) or REST.

**Module:** `github.com/smartify-inc/hivemind-go-sdk` · **Go:** 1.26+

## Install

```bash
go get github.com/smartify-inc/hivemind-go-sdk@latest
```

Set `HIVEMIND_API_KEY`. `sk_live_…` keys use production; other keys use staging unless you override the endpoint.

## Quick start

```go
client, err := hivemind.NewClient(os.Getenv("HIVEMIND_API_KEY"))
if err != nil { /* ... */ }
defer client.Close()

session := hivemind.NewSession(client,
    hivemind.WithWorkflowID("my-workflow"),
    hivemind.WithTask("Short description of the task"),
)

ctxText, err := session.GetContext(context.Background(), 4000)
// Feed ctxText into your LLM (e.g. system message).

// After the model replies:
session.RecordResponse(context.Background(), replyText, tokensUsed)
session.End(context.Background())
```

Conversation-scoped context: `GetContextWithConversation(ctx, maxTokens, conversationID)`.

REST client: `NewClient(key, hivemind.WithTransport(hivemind.TransportREST))`. Other options include `WithEndpoint`, `WithTimeout`, `WithRetryPolicy` — see [pkg.go.dev](https://pkg.go.dev/github.com/smartify-inc/hivemind-go-sdk).

Optional OpenAI integration is behind the `openai` build tag. Tests and `mock` client: same module path, subpackage `mock`.

## License

Apache 2.0 — [LICENSE](LICENSE).
