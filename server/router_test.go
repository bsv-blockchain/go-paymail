package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouterParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "paymail address param",
			input:    PaymailAddressParamName,
			expected: ":paymailAddress",
		},
		{
			name:     "pubkey param",
			input:    PubKeyParamName,
			expected: ":pubKey",
		},
		{
			name:     "custom param",
			input:    "customParam",
			expected: ":customParam",
		},
		{
			name:     "empty param",
			input:    "",
			expected: ":",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := _routerParam(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTemplateToRouterPath(t *testing.T) {
	t.Parallel()

	// Create a minimal configuration for testing
	config := &Configuration{
		APIVersion:  DefaultAPIVersion,
		ServiceName: "bsvalias",
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "paymail address template",
			template: PaymailAddressTemplate,
			expected: "/v1/bsvalias/:paymailAddress",
		},
		{
			name:     "pubkey template",
			template: PubKeyTemplate,
			expected: "/v1/bsvalias/:pubKey",
		},
		{
			name:     "path with paymail address",
			template: "address/" + PaymailAddressTemplate,
			expected: "/v1/bsvalias/address/:paymailAddress",
		},
		{
			name:     "path with pubkey",
			template: "verify/" + PaymailAddressTemplate + "/" + PubKeyTemplate,
			expected: "/v1/bsvalias/verify/:paymailAddress/:pubKey",
		},
		{
			name:     "path with leading slash",
			template: "/profile/" + PaymailAddressTemplate,
			expected: "/v1/bsvalias/profile/:paymailAddress",
		},
		{
			name:     "simple path without templates",
			template: "capabilities",
			expected: "/v1/bsvalias/capabilities",
		},
		{
			name:     "empty template",
			template: "",
			expected: "/v1/bsvalias/",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := config.templateToRouterPath(tc.template)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTemplateToRouterPath_DifferentAPIVersions(t *testing.T) {
	t.Parallel()

	t.Run("custom api version", func(t *testing.T) {
		config := &Configuration{
			APIVersion:  "v2",
			ServiceName: "bsvalias",
		}

		result := config.templateToRouterPath("test")
		assert.Equal(t, "/v2/bsvalias/test", result)
	})

	t.Run("custom service name", func(t *testing.T) {
		config := &Configuration{
			APIVersion:  "v1",
			ServiceName: "paymail",
		}

		result := config.templateToRouterPath("test")
		assert.Equal(t, "/v1/paymail/test", result)
	})
}
