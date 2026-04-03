package hivemind

import (
	"log/slog"
	"strings"
	"time"
)

type Transport int

const (
	TransportGRPC Transport = iota
	TransportREST
)

type clientConfig struct {
	apiKey      string
	endpoint    string
	transport   Transport
	hivemindID  string
	timeout     time.Duration
	retryPolicy RetryPolicy
	logger      *slog.Logger
	userAgent   string
}

type Option func(*clientConfig)

func WithEndpoint(endpoint string) Option {
	return func(c *clientConfig) { c.endpoint = endpoint }
}

func WithTransport(t Transport) Option {
	return func(c *clientConfig) { c.transport = t }
}

func WithHivemindID(id string) Option {
	return func(c *clientConfig) { c.hivemindID = id }
}

func WithTimeout(d time.Duration) Option {
	return func(c *clientConfig) { c.timeout = d }
}

func WithRetryPolicy(policy RetryPolicy) Option {
	return func(c *clientConfig) { c.retryPolicy = policy }
}

func WithLogger(handler slog.Handler) Option {
	return func(c *clientConfig) { c.logger = slog.New(handler) }
}

func WithUserAgent(ua string) Option {
	return func(c *clientConfig) { c.userAgent = ua }
}

func detectEndpoint(apiKey string, transport Transport) string {
	isLive := strings.HasPrefix(apiKey, "sk_live_")
	if transport == TransportGRPC {
		if isLive {
			return "grpc.smartify.ai:443"
		}
		return "grpc-staging.smartify.ai:443"
	}
	if isLive {
		return "https://api.smartify.ai"
	}
	return "https://api-staging.smartify.ai"
}

func defaultConfig(apiKey string) clientConfig {
	return clientConfig{
		apiKey:      apiKey,
		transport:   TransportGRPC,
		timeout:     30 * time.Second,
		retryPolicy: DefaultRetryPolicy(),
		userAgent:   "hivemind-go/0.1.0",
	}
}
