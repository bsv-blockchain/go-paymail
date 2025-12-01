package paymail

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bitcoin-sv/go-sdk/script"
)

// ResolutionResponse is the response from the ResolveAddress() request
type ResolutionResponse struct {
	StandardResponse
	ResolutionPayload
}

// ResolutionPayload is the payload from the response
type ResolutionPayload struct {
	Address   string `json:"address,omitempty"`   // Legacy BSV address derived from the output script (custom for our Go package)
	Output    string `json:"output"`              // hex-encoded Bitcoin script, which the sender MUST use during the construction of a payment transaction
	Signature string `json:"signature,omitempty"` // This is used if SenderValidation is enforced (signature of "output" value)
}

// ResolveAddress will return a hex-encoded Bitcoin script if successful
//
// Specs: http://bsvalias.org/04-01-basic-address-resolution.html
func (c *Client) ResolveAddress(resolutionURL, alias, domain string, senderRequest *SenderRequest) (response *ResolutionResponse, err error) {
	// Require a valid url
	if len(resolutionURL) == 0 || !strings.Contains(resolutionURL, "https://") {
		err = fmt.Errorf("invalid url: %s", resolutionURL)
		return response, err
	}

	// Basic requirements for the request
	if len(alias) == 0 {
		err = errors.New("missing alias")
		return response, err
	} else if len(domain) == 0 {
		err = errors.New("missing domain")
		return response, err
	}

	// Basic requirements for request
	if senderRequest == nil {
		err = errors.New("senderRequest cannot be nil")
		return response, err
	} else if len(senderRequest.Dt) == 0 {
		err = errors.New("time is required on senderRequest")
		return response, err
	} else if len(senderRequest.SenderHandle) == 0 {
		err = errors.New("sender handle is required on senderRequest")
		return response, err
	}

	// Set the base url and path, assuming the url is from the prior GetCapabilities() request
	// https://<host-discovery-target>/{alias}@{domain.tld}/payment-destination
	reqURL := replaceAliasDomain(resolutionURL, alias, domain)

	// Fire the POST request
	var resp StandardResponse
	if resp, err = c.postRequest(reqURL, senderRequest); err != nil {
		return response, err
	}

	// Start the response
	response = &ResolutionResponse{StandardResponse: resp}

	// Test the status code
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotModified {

		// Paymail address not found?
		if response.StatusCode == http.StatusNotFound {
			err = errors.New("paymail address not found")
		} else {
			serverError := &ServerError{}
			if err = json.Unmarshal(resp.Body, serverError); err != nil {
				return response, err
			}
			err = fmt.Errorf("bad response from paymail provider: code %d, message: %s", response.StatusCode, serverError.Message)
		}

		return response, err
	}

	// Decode the body of the response
	if err = json.Unmarshal(resp.Body, &response); err != nil {
		return response, err
	}

	// Check for an output
	if len(response.Output) == 0 {
		err = errors.New("missing an output value")
		return response, err
	}

	script, err := script.NewFromHex(response.Output)
	if err != nil {
		return response, err
	}

	addresses, err := script.Addresses()
	if err != nil || len(addresses) == 0 {
		err = errors.New("invalid output script, missing an address")
		return response, err
	}

	response.Address = addresses[0]

	return response, err
}
