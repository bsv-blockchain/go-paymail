package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test_getSenderPubKey will test the method getSenderPubKey()
func Test_getSenderPubKey(t *testing.T) {
	// todo: this needs proper mocking

	t.Run("error - bad domain", func(t *testing.T) {
		key, err := getSenderPubKey("bad@domain.com")
		require.Error(t, err)
		require.Nil(t, key)
	})

	t.Run("valid - good paymail", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration testing in short mode")
		}
		key, err := getSenderPubKey("mrzz@handcash.io")
		if err != nil {
			t.Skipf("skipping: external paymail service unavailable: %s", err.Error())
		}
		require.NotNil(t, key)
	})
}
