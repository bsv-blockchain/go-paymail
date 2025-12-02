package paymail

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

/*
Default Response:
{
  "bsvalias": "1.0",
  "handle": "<alias>@<domain>.<tld>",
  "pubkey": "..."
}
*/

var (
	// ErrPKIInvalidURL is returned when URL is invalid
	ErrPKIInvalidURL = errors.New("invalid url")
	// ErrPKIMissingBsvAlias is returned when bsvalias version is missing
	ErrPKIMissingBsvAlias = errors.New("missing bsvalias version")
	// ErrPKIBadResponse is returned when paymail provider returns bad response
	ErrPKIBadResponse = errors.New("bad response from paymail provider")
	// ErrPKIHandleMismatch is returned when handle does not match paymail address
	ErrPKIHandleMismatch = errors.New("pki response handle does not match paymail address")
	// ErrPKIMissingPubKey is returned when PubKey is missing
	ErrPKIMissingPubKey = errors.New("pki response is missing a PubKey value")
	// ErrPKIInvalidPubKeyLength is returned when PubKey length is invalid
	ErrPKIInvalidPubKeyLength = errors.New("returned pubkey is not the required length")
)

// PKIResponse is the result returned
type PKIResponse struct {
	StandardResponse
	PKIPayload
}

// PKIPayload is the payload from the response
type PKIPayload struct {
	BsvAlias string `json:"bsvalias"` // Version of Paymail
	Handle   string `json:"handle"`   // The <alias>@<domain>.<tld>
	PubKey   string `json:"pubkey"`   // The related PubKey
}

// GetPKI will return a valid PKI response for a given alias@domain.tld
//
// Specs: http://bsvalias.org/03-public-key-infrastructure.html
func (c *Client) GetPKI(pkiURL, alias, domain string) (response *PKIResponse, err error) {
	// Require a valid url
	if len(pkiURL) == 0 || !strings.Contains(pkiURL, "https://") {
		err = fmt.Errorf("url %s: %w", pkiURL, ErrPKIInvalidURL)
		return response, err
	}

	// Basic requirements for the request
	if len(alias) == 0 {
		err = ErrPikeMissingAlias
		return response, err
	} else if len(domain) == 0 {
		err = ErrPikeMissingDomain
		return response, err
	}

	// Set the base url and path, assuming the url is from the prior GetCapabilities() request
	// https://<host-discovery-target>/{alias}@{domain.tld}/id
	reqURL := replaceAliasDomain(pkiURL, alias, domain)

	// Fire the GET request
	var resp StandardResponse
	if resp, err = c.getRequest(reqURL); err != nil {
		return response, err
	}

	// Start the response
	response = &PKIResponse{StandardResponse: resp}

	// Test the status code (200 or 304 is valid)
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotModified {
		serverError := &ServerError{}
		if err = json.Unmarshal(resp.Body, serverError); err != nil {
			return response, err
		}
		err = fmt.Errorf("code %d, message: %s: %w", response.StatusCode, serverError.Message, ErrPKIBadResponse)
		return response, err
	}

	// Decode the body of the response
	if err = json.Unmarshal(resp.Body, &response); err != nil {
		return response, err
	}

	// Invalid version detected
	if len(response.BsvAlias) == 0 {
		err = ErrPKIMissingBsvAlias
		return response, err
	}

	// Check basic requirements (handle should match our alias@domain.tld)
	if response.Handle != alias+"@"+domain {
		err = fmt.Errorf("handle %s vs %s: %w", response.Handle, alias+"@"+domain, ErrPKIHandleMismatch)
		return response, err
	}

	// Check the PubKey length
	if len(response.PubKey) == 0 {
		err = ErrPKIMissingPubKey
	} else if len(response.PubKey) != PubKeyLength {
		err = fmt.Errorf("expected length %d, got: %d: %w", PubKeyLength, len(response.PubKey), ErrPKIInvalidPubKeyLength)
	}

	return response, err
}
