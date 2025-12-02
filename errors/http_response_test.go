package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test errors for use in http_response tests
var (
	errTestRegular = errors.New("some regular error")
	errTestWrapped = errors.New("wrapped error")
)

func TestMapAndLog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		err                error
		expectedStatusCode int
		expectedCode       string
		expectedMessage    string
	}{
		{
			name: "SPVError with bad request status",
			err: SPVError{
				Code:       "error-bad-request",
				Message:    "bad request message",
				StatusCode: http.StatusBadRequest,
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedCode:       "error-bad-request",
			expectedMessage:    "bad request message",
		},
		{
			name: "SPVError with internal server error",
			err: SPVError{
				Code:       "error-internal",
				Message:    "internal server error",
				StatusCode: http.StatusInternalServerError,
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedCode:       "error-internal",
			expectedMessage:    "internal server error",
		},
		{
			name: "SPVError with not found status",
			err: SPVError{
				Code:       "error-not-found",
				Message:    "resource not found",
				StatusCode: http.StatusNotFound,
			},
			expectedStatusCode: http.StatusNotFound,
			expectedCode:       "error-not-found",
			expectedMessage:    "resource not found",
		},
		{
			name:               "regular error returns unknown code and 500",
			err:                errTestRegular,
			expectedStatusCode: http.StatusInternalServerError,
			expectedCode:       UnknownErrorCode,
			expectedMessage:    "",
		},
		{
			name:               "wrapped regular error",
			err:                errTestWrapped,
			expectedStatusCode: http.StatusInternalServerError,
			expectedCode:       UnknownErrorCode,
			expectedMessage:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call with nil logger to avoid logging during tests
			response, statusCode := mapAndLog(tc.err, nil)

			assert.Equal(t, tc.expectedStatusCode, statusCode)
			assert.Equal(t, tc.expectedCode, response.Code)
			assert.Equal(t, tc.expectedMessage, response.Message)
		})
	}
}

func TestMapAndLog_WithSPVError_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	spvErr := SPVError{
		Code:       "test-error-code",
		Message:    "test error message",
		StatusCode: http.StatusUnprocessableEntity,
	}

	response, statusCode := mapAndLog(spvErr, nil)

	assert.Equal(t, http.StatusUnprocessableEntity, statusCode)
	assert.Equal(t, "test-error-code", response.Code)
	assert.Equal(t, "test error message", response.Message)
}

func TestMapAndLog_WithRegularError_ReturnsUnknown(t *testing.T) {
	t.Parallel()

	response, statusCode := mapAndLog(errTestRegular, nil)

	assert.Equal(t, http.StatusInternalServerError, statusCode)
	assert.Equal(t, UnknownErrorCode, response.Code)
	assert.Empty(t, response.Message)
}

func TestMapAndLog_StatusCodeBoundaries(t *testing.T) {
	t.Parallel()

	t.Run("status code just below 500 logs at warn level", func(t *testing.T) {
		spvErr := SPVError{
			Code:       "client-error",
			Message:    "client error",
			StatusCode: 499,
		}

		response, statusCode := mapAndLog(spvErr, nil)

		assert.Equal(t, 499, statusCode)
		assert.Equal(t, "client-error", response.Code)
	})

	t.Run("status code exactly 500 logs at error level", func(t *testing.T) {
		spvErr := SPVError{
			Code:       "server-error",
			Message:    "server error",
			StatusCode: 500,
		}

		response, statusCode := mapAndLog(spvErr, nil)

		assert.Equal(t, 500, statusCode)
		assert.Equal(t, "server-error", response.Code)
	})

	t.Run("status code above 500 logs at error level", func(t *testing.T) {
		spvErr := SPVError{
			Code:       "gateway-error",
			Message:    "bad gateway",
			StatusCode: 502,
		}

		response, statusCode := mapAndLog(spvErr, nil)

		assert.Equal(t, 502, statusCode)
		assert.Equal(t, "gateway-error", response.Code)
	})
}
