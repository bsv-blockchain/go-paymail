package paymail

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tests for client functional options (client_options.go)
// TestDefaultClientOptions is already covered in client_test.go

func TestWithDNSPort(t *testing.T) {
	t.Parallel()

	opts := &ClientOptions{}
	WithDNSPort("5353")(opts)

	assert.Equal(t, "5353", opts.dnsPort)
}

func TestWithDNSTimeout(t *testing.T) {
	t.Parallel()

	opts := &ClientOptions{}
	WithDNSTimeout(30 * time.Second)(opts)

	assert.Equal(t, 30*time.Second, opts.dnsTimeout)
}

func TestWithBRFCSpecs(t *testing.T) {
	t.Parallel()

	specs := []*BRFCSpec{
		{ID: "test-spec-1"},
		{ID: "test-spec-2"},
	}

	opts := &ClientOptions{}
	WithBRFCSpecs(specs)(opts)

	assert.Equal(t, specs, opts.brfcSpecs)
	assert.Len(t, opts.brfcSpecs, 2)
}

func TestWithHTTPTimeout(t *testing.T) {
	t.Parallel()

	opts := &ClientOptions{}
	WithHTTPTimeout(60 * time.Second)(opts)

	assert.Equal(t, 60*time.Second, opts.httpTimeout)
}

func TestWithNameServer(t *testing.T) {
	t.Parallel()

	opts := &ClientOptions{}
	WithNameServer("1.1.1.1")(opts)

	assert.Equal(t, "1.1.1.1", opts.nameServer)
}

func TestWithNameServerNetwork(t *testing.T) {
	t.Parallel()

	opts := &ClientOptions{}
	WithNameServerNetwork("tcp")(opts)

	assert.Equal(t, "tcp", opts.nameServerNetwork)
}

func TestWithRequestTracing(t *testing.T) {
	t.Parallel()

	opts := &ClientOptions{}
	assert.False(t, opts.requestTracing)

	WithRequestTracing()(opts)

	assert.True(t, opts.requestTracing)
}

func TestWithRetryCount(t *testing.T) {
	t.Parallel()

	opts := &ClientOptions{}
	WithRetryCount(5)(opts)

	assert.Equal(t, 5, opts.retryCount)
}

func TestWithSSLTimeout(t *testing.T) {
	t.Parallel()

	opts := &ClientOptions{}
	WithSSLTimeout(30 * time.Second)(opts)

	assert.Equal(t, 30*time.Second, opts.sslTimeout)
}

func TestWithSSLDeadline(t *testing.T) {
	t.Parallel()

	opts := &ClientOptions{}
	WithSSLDeadline(15 * time.Second)(opts)

	assert.Equal(t, 15*time.Second, opts.sslDeadline)
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	opts := &ClientOptions{}
	WithUserAgent("custom-agent/1.0")(opts)

	assert.Equal(t, "custom-agent/1.0", opts.userAgent)
}

func TestWithNetwork(t *testing.T) {
	t.Parallel()

	t.Run("set mainnet", func(t *testing.T) {
		opts := &ClientOptions{}
		WithNetwork(Mainnet)(opts)

		assert.Equal(t, Mainnet, opts.network)
	})

	t.Run("set testnet", func(t *testing.T) {
		opts := &ClientOptions{}
		WithNetwork(Testnet)(opts)

		assert.Equal(t, Testnet, opts.network)
	})

	t.Run("set STN", func(t *testing.T) {
		opts := &ClientOptions{}
		WithNetwork(STN)(opts)

		assert.Equal(t, STN, opts.network)
	})
}

func TestClientOptions_ChainedOptions(t *testing.T) {
	t.Parallel()

	opts := &ClientOptions{}

	// Apply multiple options
	WithDNSPort("5353")(opts)
	WithDNSTimeout(30 * time.Second)(opts)
	WithHTTPTimeout(60 * time.Second)(opts)
	WithNameServer("1.1.1.1")(opts)
	WithRetryCount(5)(opts)
	WithRequestTracing()(opts)
	WithUserAgent("test-agent")(opts)
	WithNetwork(Testnet)(opts)

	// Verify all values
	assert.Equal(t, "5353", opts.dnsPort)
	assert.Equal(t, 30*time.Second, opts.dnsTimeout)
	assert.Equal(t, 60*time.Second, opts.httpTimeout)
	assert.Equal(t, "1.1.1.1", opts.nameServer)
	assert.Equal(t, 5, opts.retryCount)
	assert.True(t, opts.requestTracing)
	assert.Equal(t, "test-agent", opts.userAgent)
	assert.Equal(t, Testnet, opts.network)
}
