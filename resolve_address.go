package paymail

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bsv-blockchain/go-sdk/script"
)

var (
	// ErrResolveAddressInvalidURL is returned when resolution URL is invalid
	ErrResolveAddressInvalidURL = errors.New("invalid url")
	// ErrResolveAddressMissingAlias is returned when alias is missing
	ErrResolveAddressMissingAlias = errors.New("missing alias")
	// ErrResolveAddressMissingDomain is returned when domain is missing
	ErrResolveAddressMissingDomain = errors.New("missing domain")
	// ErrResolveAddressNilRequest is returned when sender request is nil
	ErrResolveAddressNilRequest = errors.New("senderRequest cannot be nil")
	// ErrResolveAddressMissingDt is returned when dt is missing from request
	ErrResolveAddressMissingDt = errors.New("time is required on senderRequest")
	// ErrResolveAddressMissingSenderHandle is returned when sender handle is missing from request
	ErrResolveAddressMissingSenderHandle = errors.New("sender handle is required on senderRequest")
	// ErrResolveAddressNotFound is returned when paymail address is not found
	ErrResolveAddressNotFound = errors.New("paymail address not found")
	// ErrResolveAddressBadResponse is returned when receiving bad response from paymail provider
	ErrResolveAddressBadResponse = errors.New("bad response from paymail provider")
	// ErrResolveAddressMissingOutput is returned when output is missing from response
	ErrResolveAddressMissingOutput = errors.New("missing an output value")
	// ErrResolveAddressInvalidScript is returned when output script is invalid
	ErrResolveAddressInvalidScript = errors.New("invalid output script, missing an address")
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
		err = fmt.Errorf("%s: %s: %w", "invalid url", resolutionURL, ErrResolveAddressInvalidURL)
		return response, err
	}

	// Basic requirements for the request
	if len(alias) == 0 {
		err = ErrResolveAddressMissingAlias
		return response, err
	} else if len(domain) == 0 {
		err = ErrResolveAddressMissingDomain
		return response, err
	}

	// Basic requirements for request
	if senderRequest == nil {
		err = ErrResolveAddressNilRequest
		return response, err
	} else if len(senderRequest.Dt) == 0 {
		err = ErrResolveAddressMissingDt
		return response, err
	} else if len(senderRequest.SenderHandle) == 0 {
		err = ErrResolveAddressMissingSenderHandle
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
			err = ErrResolveAddressNotFound
		} else {
			serverError := &ServerError{}
			if err = json.Unmarshal(resp.Body, serverError); err != nil {
				return response, err
			}
			err = fmt.Errorf("code %d, message: %s: %w", response.StatusCode, serverError.Message, ErrResolveAddressBadResponse)
		}

		return response, err
	}

	// Decode the body of the response
	if err = json.Unmarshal(resp.Body, &response); err != nil {
		return response, err
	}

	// Check for an output
	if len(response.Output) == 0 {
		err = ErrResolveAddressMissingOutput
		return response, err
	}

	script, err := script.NewFromHex(response.Output)
	if err != nil {
		return response, err
	}

	addresses, err := script.Addresses()
	if err != nil || len(addresses) == 0 {
		err = ErrResolveAddressInvalidScript
		return response, err
	}

	response.Address = addresses[0]

	return response, err
}
