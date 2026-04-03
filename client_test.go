package hivemind_test

import (
	"testing"
	"time"

	hivemind "github.com/smartifyai/hivemind-go"
)

func TestNewClientRequiresApiKey(t *testing.T) {
	_, err := hivemind.NewClient("")
	if err == nil {
		t.Fatalf("expected error for empty API key, got nil")
	}
}

func TestNewClientAcceptsValidKey(t *testing.T) {
	c, err := hivemind.NewClient("sk_test_xxx", hivemind.WithTransport(hivemind.TransportREST))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer c.Close()
}

func TestEndpointDetectionLiveGRPC(t *testing.T) {
	got := hivemind.DetectEndpoint("sk_live_xxx", hivemind.TransportGRPC)
	want := "grpc.smartify.ai:443"
	if got != want {
		t.Errorf("DetectEndpoint(live, gRPC) = %q, want %q", got, want)
	}
}

func TestEndpointDetectionLiveREST(t *testing.T) {
	got := hivemind.DetectEndpoint("sk_live_xxx", hivemind.TransportREST)
	want := "https://api.smartify.ai"
	if got != want {
		t.Errorf("DetectEndpoint(live, REST) = %q, want %q", got, want)
	}
}

func TestEndpointDetectionTestGRPC(t *testing.T) {
	got := hivemind.DetectEndpoint("sk_test_xxx", hivemind.TransportGRPC)
	want := "grpc-staging.smartify.ai:443"
	if got != want {
		t.Errorf("DetectEndpoint(test, gRPC) = %q, want %q", got, want)
	}
}

func TestEndpointDetectionTestREST(t *testing.T) {
	got := hivemind.DetectEndpoint("sk_test_xxx", hivemind.TransportREST)
	want := "https://api-staging.smartify.ai"
	if got != want {
		t.Errorf("DetectEndpoint(test, REST) = %q, want %q", got, want)
	}
}

func TestWithEndpointOverride(t *testing.T) {
	custom := "https://custom.example.com"
	c, err := hivemind.NewClient("sk_test_xxx",
		hivemind.WithTransport(hivemind.TransportREST),
		hivemind.WithEndpoint(custom),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer c.Close()
}

func TestDefaultConfig(t *testing.T) {
	apiKey, transport, timeout, userAgent, _ := hivemind.GetDefaultConfig("sk_test_abc")

	if apiKey != "sk_test_abc" {
		t.Errorf("apiKey = %q, want %q", apiKey, "sk_test_abc")
	}
	if transport != hivemind.TransportGRPC {
		t.Errorf("transport = %v, want TransportGRPC", transport)
	}
	if timeout != 30*time.Second {
		t.Errorf("timeout = %v, want 30s", timeout)
	}
	if userAgent != "hivemind-go/0.1.0" {
		t.Errorf("userAgent = %q, want %q", userAgent, "hivemind-go/0.1.0")
	}
}
