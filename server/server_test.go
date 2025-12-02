package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateServer will test the method CreateServer()
func TestCreateServer(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := testConfig(t, "localhost")
		config.Port = 12345
		config.Timeout = 10 * time.Second
		s := CreateServer(config)
		require.NotNil(t, s)
		//nolint:gosec // G112: This is a type assertion test, not server creation
		assert.IsType(t, &http.Server{}, s)
		assert.Equal(t, fmt.Sprintf(":%d", config.Port), s.Addr)
		assert.Equal(t, config.Timeout, s.WriteTimeout)
		assert.Equal(t, config.Timeout, s.ReadTimeout)
	})
}

// TestWithServer will test if the server is running and responding to capabilities discovery & each capability is accessible
func TestWithServer(t *testing.T) {
	t.Run("run server and check capabilities", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))

		config, _ := NewConfig(sl, WithDomain("domain.com"))
		config.Prefix = "http://"

		server := httptest.NewServer(Handlers(config))
		defer server.Close()

		err := config.AddDomain(server.URL)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/.well-known/bsvalias", server.URL), nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make GET request: %v", err)
		}

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, result["bsvalias"], config.BSVAliasVersion)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		require.NoError(t, resp.Body.Close())

		capabilities := result["capabilities"].(map[string]interface{})
		assert.NotNil(t, capabilities)
		assert.NotEmpty(t, capabilities)

		// Check if all callable capabilities are accessible by trying to make a request to each one of them
		for _, cap := range capabilities {
			capUrl, ok := cap.(string)
			if !ok {
				continue // skip static capabilities
			}

			capUrl = strings.ReplaceAll(capUrl, PaymailAddressTemplate, "example@domain.com")
			capUrl = strings.ReplaceAll(capUrl, PubKeyTemplate, "xpub")

			_, err := url.Parse(capUrl)
			require.NoError(t, err, "Endpoint %s is not a valid URL", capUrl)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, capUrl, nil)
			if err != nil {
				t.Logf("Failed to create request: %v", err)
				continue
			}
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				_ = resp.Body.Close()
			}

			// Only verify if the current 'capUrl' endpoint is accessible, even if the 'GET' method is not permitted for it.
			assert.NoError(t, err)
		}
	})
}
