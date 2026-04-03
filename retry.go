package hivemind

import (
	"math/rand"
	"time"
)

type RetryPolicy struct {
	MaxRetries           int           `json:"max_retries"`
	InitialBackoff       time.Duration `json:"initial_backoff"`
	MaxBackoff           time.Duration `json:"max_backoff"`
	BackoffFactor        float64       `json:"backoff_factor"`
	RetryableStatusCodes []int         `json:"retryable_status_codes"`
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:           3,
		InitialBackoff:       100 * time.Millisecond,
		MaxBackoff:           5 * time.Second,
		BackoffFactor:        2.0,
		RetryableStatusCodes: []int{429, 502, 503, 504},
	}
}

func (p RetryPolicy) isRetryable(statusCode int) bool {
	for _, code := range p.RetryableStatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

func (p RetryPolicy) backoff(attempt int) time.Duration {
	d := float64(p.InitialBackoff)
	for i := 0; i < attempt; i++ {
		d *= p.BackoffFactor
	}
	if time.Duration(d) > p.MaxBackoff {
		d = float64(p.MaxBackoff)
	}
	jitter := d * 0.1 * (rand.Float64()*2 - 1)
	return time.Duration(d + jitter)
}
