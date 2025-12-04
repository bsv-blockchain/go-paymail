package paymail

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

// ExampleClient_GetOutputsTemplate example using GetOutputsTemplate()
//
// See more examples in /examples/
func ExampleClient_GetOutputsTemplate() {
	// Activate httpmock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Setup a mock HTTP client
	client := newTestClient(nil)
	var amount uint64 = 1000
	mockPIKEOutputs(http.StatusOK, amount)

	// Assume we have a PIKE Outputs URL
	pikeOutputsURL := "https://test.com/v1/bsvalias/pike/outputs/{alias}@{domain.tld}"

	// Prepare the payload
	payload := &PikePaymentOutputsPayload{
		SenderPaymail: "joedoe@example.com",
		Amount:        amount, // Example amount in satoshis
	}

	// Get the outputs template from PIKE
	outputs, err := client.GetOutputsTemplate(pikeOutputsURL, "alias", "domain.tld", payload)
	if err != nil {
		fmt.Printf("error getting outputs template: %s", err.Error())
		return
	}
	fmt.Printf("found outputs template, reference: %+v", outputs.Reference)
	// Output: found outputs template, reference: 1262077636c27af74c01bb4535a7a90e
}

// TestClient_GetOutputsTemplate will test the method GetOutputsTemplate()
func TestClient_GetOutputsTemplate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := newTestClient(t)

	t.Run("successful PIKE outputs response", func(t *testing.T) {
		var amount uint64 = 1000
		mockPIKEOutputs(http.StatusOK, amount)

		outputsURL := "https://" + testDomain + "/v1/bsvalias/pike/outputs/{alias}@{domain.tld}"
		payload := &PikePaymentOutputsPayload{
			SenderPaymail: "joedoe@example.com",
			Amount:        amount,
		}
		response, err := client.GetOutputsTemplate(outputsURL, "alias", "domain.tld", payload)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.NotNil(t, response.Reference)
		require.Equal(t, payload.Amount, response.Outputs[0].Satoshis)
	})

	t.Run("PIKE outputs response error", func(t *testing.T) {
		httpmock.Reset()
		httpmock.RegisterResponder(http.MethodPost, "https://example.com/v1/bsvalias/pike/outputs/alias@domain.tld",
			httpmock.NewStringResponder(http.StatusBadRequest, `{"message": "bad request"}`),
		)

		outputsURL := "https://example.com/v1/bsvalias/pike/outputs/{alias}@{domain.tld}"
		payload := &PikePaymentOutputsPayload{
			SenderPaymail: "joedoe@example.com",
			Amount:        1000,
		}
		response, err := client.GetOutputsTemplate(outputsURL, "alias", "domain.tld", payload)
		require.Error(t, err)
		require.Nil(t, response)
	})
}

// mockPIKEOutputs is used for mocking the PIKE outputs response
func mockPIKEOutputs(statusCode int, amount uint64) {
	httpmock.RegisterResponder(http.MethodPost, "https://"+testDomain+"/v1/bsvalias/pike/outputs/alias@domain.tld",
		httpmock.NewStringResponder(
			statusCode,
			`{
                "reference": "1262077636c27af74c01bb4535a7a90e",
				"outputs": [
					{
						"script": "76a9fd88ac",
						"satoshis": `+strconv.Itoa(int(amount))+`
					}
				]
            }`,
		),
	)
}

// BenchmarkClient_GetOutputsTemplate benchmarks the method GetOutputsTemplate()
func BenchmarkClient_GetOutputsTemplate(b *testing.B) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := newTestClient(nil)
	mockPIKEOutputs(http.StatusOK, 1000)

	outputsURL := "https://example.com/v1/bsvalias/pike/outputs/{alias}@{domain.tld}"
	payload := &PikePaymentOutputsPayload{
		SenderPaymail: "joedoe@example.com",
		Amount:        1000, // Example amount in satoshis
	}

	for i := 0; i < b.N; i++ {
		_, _ = client.GetOutputsTemplate(outputsURL, "alias", "domain.tld", payload)
	}
}
