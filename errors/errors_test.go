package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSPVError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      SPVError
		expected string
	}{
		{
			name:     "returns message",
			err:      SPVError{Code: "test-code", Message: "test message", StatusCode: 400},
			expected: "test message",
		},
		{
			name:     "empty message",
			err:      SPVError{Code: "test-code", Message: "", StatusCode: 500},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.err.Error()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSPVError_GetCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      SPVError
		expected string
	}{
		{
			name:     "returns code",
			err:      SPVError{Code: "error-invalid-input", Message: "msg", StatusCode: 400},
			expected: "error-invalid-input",
		},
		{
			name:     "empty code",
			err:      SPVError{Code: "", Message: "msg", StatusCode: 500},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.err.GetCode()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSPVError_GetMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      SPVError
		expected string
	}{
		{
			name:     "returns message",
			err:      SPVError{Code: "code", Message: "error message", StatusCode: 400},
			expected: "error message",
		},
		{
			name:     "empty message",
			err:      SPVError{Code: "code", Message: "", StatusCode: 400},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.err.GetMessage()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSPVError_GetStatusCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      SPVError
		expected int
	}{
		{
			name:     "returns bad request status",
			err:      SPVError{Code: "code", Message: "msg", StatusCode: http.StatusBadRequest},
			expected: http.StatusBadRequest,
		},
		{
			name:     "returns internal server error status",
			err:      SPVError{Code: "code", Message: "msg", StatusCode: http.StatusInternalServerError},
			expected: http.StatusInternalServerError,
		},
		{
			name:     "returns not found status",
			err:      SPVError{Code: "code", Message: "msg", StatusCode: http.StatusNotFound},
			expected: http.StatusNotFound,
		},
		{
			name:     "zero status code",
			err:      SPVError{Code: "code", Message: "msg", StatusCode: 0},
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.err.GetStatusCode()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSPVError_ImplementsErrorInterface(t *testing.T) {
	t.Parallel()

	// Verify SPVError implements the error interface
	var e error = SPVError{Code: "test", Message: "test message", StatusCode: 400}
	require.Error(t, e)
	assert.Equal(t, "test message", e.Error())
}

func TestSPVError_ImplementsExtendedErrorInterface(t *testing.T) {
	t.Parallel()

	var extErr ExtendedError = SPVError{Code: "test", Message: "test message", StatusCode: 400}
	assert.NotNil(t, extErr)
	assert.Equal(t, "test", extErr.GetCode())
	assert.Equal(t, "test message", extErr.GetMessage())
	assert.Equal(t, 400, extErr.GetStatusCode())
}

func TestResponseError(t *testing.T) {
	t.Parallel()

	respErr := ResponseError{Code: "error-code", Message: "error message"}
	assert.Equal(t, "error-code", respErr.Code)
	assert.Equal(t, "error message", respErr.Message)
}

func TestUnknownErrorCode(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "error-unknown", UnknownErrorCode)
}
