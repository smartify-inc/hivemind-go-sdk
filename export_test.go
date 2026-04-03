package hivemind

import "time"

var DetectEndpoint = detectEndpoint

func GetDefaultConfig(apiKey string) (string, Transport, time.Duration, string, RetryPolicy) {
	cfg := defaultConfig(apiKey)
	return cfg.apiKey, cfg.transport, cfg.timeout, cfg.userAgent, cfg.retryPolicy
}

func (p RetryPolicy) IsRetryable(statusCode int) bool {
	return p.isRetryable(statusCode)
}

func (p RetryPolicy) Backoff(attempt int) time.Duration {
	return p.backoff(attempt)
}
