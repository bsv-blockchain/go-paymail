package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreateMetadata will test the method CreateMetadata()
func TestCreateMetadata(t *testing.T) {
	t.Parallel()

	t.Run("invalid empty request", func(t *testing.T) {
		req := new(http.Request)
		md := CreateMetadata(req, "tester", "test.com", "optional")
		assert.NotNil(t, md)
		assert.Equal(t, "tester", md.Alias)
		assert.Equal(t, "test.com", md.Domain)
		assert.Equal(t, "optional", md.Note)
		assert.Empty(t, md.UserAgent)
		assert.Empty(t, md.RequestURI)
		assert.Empty(t, md.IPAddress)
		assert.Nil(t, md.ResolveAddress)
		assert.Nil(t, md.PaymentDestination)
	})

	// todo: add more tests on parsing request for IP, user agent etc
}
