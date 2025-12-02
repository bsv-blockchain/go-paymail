package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateServiceURL(t *testing.T) {
	t.Parallel()

	t.Run("valid values", func(t *testing.T) {
		u, err := generateServiceURL("https://", "test.com", "v1", "bsvalias")
		require.NoError(t, err)
		assert.Equal(t, "https://test.com/v1/bsvalias", u)
	})

	t.Run("all invalid values", func(t *testing.T) {
		_, err := generateServiceURL("", "", "", "")
		assert.Error(t, err)
	})

	t.Run("missing prefix", func(t *testing.T) {
		_, err := generateServiceURL("", "test.com", "v1", "")
		assert.Error(t, err)
	})

	t.Run("missing domain", func(t *testing.T) {
		_, err := generateServiceURL("https://", "", "v1", "")
		assert.Error(t, err)
	})

	t.Run("no api version", func(t *testing.T) {
		u, err := generateServiceURL("https://", "test", "", "bsvalias")
		require.NoError(t, err)
		assert.Equal(t, "https://test/bsvalias", u)
	})

	t.Run("no service name", func(t *testing.T) {
		u, err := generateServiceURL("https://", "test", "v1", "")
		require.NoError(t, err)
		assert.Equal(t, "https://test/v1", u)
	})

	t.Run("service with explicit port", func(t *testing.T) {
		u, err := generateServiceURL("https://", "test:1234", "v1", "bsvalias")
		require.NoError(t, err)
		assert.Equal(t, "https://test:1234/v1/bsvalias", u)
	})
}
