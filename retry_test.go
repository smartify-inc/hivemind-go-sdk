package hivemind_test

import (
	"testing"
	"time"

	hivemind "github.com/smartify-inc/hivemind-go-sdk"
)

func TestDefaultRetryPolicy(t *testing.T) {
	p := hivemind.DefaultRetryPolicy()

	if p.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", p.MaxRetries)
	}
	if p.InitialBackoff != 100*time.Millisecond {
		t.Errorf("InitialBackoff = %v, want 100ms", p.InitialBackoff)
	}

	want := map[int]bool{429: true, 502: true, 503: true, 504: true}
	for _, code := range p.RetryableStatusCodes {
		if !want[code] {
			t.Errorf("unexpected retryable code %d", code)
		}
		delete(want, code)
	}
	for code := range want {
		t.Errorf("missing retryable code %d", code)
	}
}

func TestIsRetryable(t *testing.T) {
	p := hivemind.DefaultRetryPolicy()

	cases := []struct {
		code int
		want bool
	}{
		{429, true},
		{502, true},
		{500, false},
		{200, false},
	}
	for _, tc := range cases {
		if got := p.IsRetryable(tc.code); got != tc.want {
			t.Errorf("IsRetryable(%d) = %v, want %v", tc.code, got, tc.want)
		}
	}
}

func TestBackoffExponential(t *testing.T) {
	p := hivemind.DefaultRetryPolicy()

	expectations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
	}
	for attempt, expected := range expectations {
		got := p.Backoff(attempt)
		low := time.Duration(float64(expected) * 0.9)
		high := time.Duration(float64(expected) * 1.1)
		if got < low || got > high {
			t.Errorf("Backoff(%d) = %v, want ~%v (range %v–%v)", attempt, got, expected, low, high)
		}
	}
}

func TestBackoffCappedAtMax(t *testing.T) {
	p := hivemind.DefaultRetryPolicy()
	got := p.Backoff(10)
	// backoff applies ±10% jitter after capping, so allow up to 10% over MaxBackoff
	limit := time.Duration(float64(p.MaxBackoff) * 1.1)
	if got > limit {
		t.Errorf("Backoff(10) = %v, exceeds MaxBackoff %v (+10%% jitter = %v)", got, p.MaxBackoff, limit)
	}
}

func Test500NotRetryedByDefault(t *testing.T) {
	p := hivemind.DefaultRetryPolicy()
	if p.IsRetryable(500) {
		t.Error("DefaultRetryPolicy should not retry 500")
	}
}
