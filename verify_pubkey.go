package paymail

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	// ErrVerifyPubKeyInvalidURL is returned when verify URL is invalid
	ErrVerifyPubKeyInvalidURL = errors.New("invalid url")
	// ErrVerifyPubKeyMissingAlias is returned when alias is missing
	ErrVerifyPubKeyMissingAlias = errors.New("missing alias")
	// ErrVerifyPubKeyMissingDomain is returned when domain is missing
	ErrVerifyPubKeyMissingDomain = errors.New("missing domain")
	// ErrVerifyPubKeyMissingPubKey is returned when pubkey is missing
	ErrVerifyPubKeyMissingPubKey = errors.New("missing pubKey")
	// ErrVerifyPubKeyBadResponse is returned when receiving bad response from paymail provider
	ErrVerifyPubKeyBadResponse = errors.New("bad response from paymail provider")
	// ErrVerifyPubKeyMissingVersion is returned when bsvalias version is missing
	ErrVerifyPubKeyMissingVersion = errors.New("missing bsvalias version")
	// ErrVerifyPubKeyMissingPubKeyValue is returned when pki response is missing a pubkey value
	ErrVerifyPubKeyMissingPubKeyValue = errors.New("pki response is missing a PubKey value")
	// ErrVerifyPubKeyInvalidLength is returned when returned pubkey is not the required length
	ErrVerifyPubKeyInvalidLength = errors.New("returned pubkey is not the required length")
	// ErrVerifyPubKeyHandleNotMatching is returned when verify response handle does not match paymail address
	ErrVerifyPubKeyHandleNotMatching = errors.New("verify response handle does not match paymail address")
)

/*
Default:

{
  "handle":"somepaymailhandle@domain.tld",
  "match": true,
  "pubkey":"<consulted pubkey>"
}
*/

// VerificationResponse is the result returned from the VerifyPubKey() request
type VerificationResponse struct {
	StandardResponse
	VerificationPayload
}

// VerificationPayload is the payload from the response
type VerificationPayload struct {
	BsvAlias string `json:"bsvalias"` // Version of the bsvalias
	Handle   string `json:"handle"`   // The <alias>@<domain>.<tld>
	Match    bool   `json:"match"`    // If the match was successful or not
	PubKey   string `json:"pubkey"`   // The related PubKey
}

// VerifyPubKey will try to match a handle and pubkey
//
// Specs: https://bsvalias.org/05-verify-public-key-owner.html
func (c *Client) VerifyPubKey(verifyURL, alias, domain, pubKey string) (response *VerificationResponse, err error) {
	// Require a valid url
	if len(verifyURL) == 0 || !strings.Contains(verifyURL, "https://") {
		err = fmt.Errorf("%s: %s: %w", "invalid url", verifyURL, ErrVerifyPubKeyInvalidURL)
		return response, err
	}

	// Basic requirements for request
	if len(alias) == 0 {
		err = ErrVerifyPubKeyMissingAlias
		return response, err
	} else if len(domain) == 0 {
		err = ErrVerifyPubKeyMissingDomain
		return response, err
	} else if len(pubKey) == 0 {
		err = ErrVerifyPubKeyMissingPubKey
		return response, err
	}

	// Set the base url and path, assuming the url is from the prior GetCapabilities() request
	// https://<host-discovery-target>/verifypubkey/{alias}@{domain.tld}/{pubkey}
	reqURL := replacePubKey(replaceAliasDomain(verifyURL, alias, domain), pubKey)

	// Fire the GET request
	var resp StandardResponse
	if resp, err = c.getRequest(reqURL); err != nil {
		return response, err
	}

	// Start the response
	response = &VerificationResponse{StandardResponse: resp}

	// Test the status code (200 or 304 is valid)
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotModified {
		serverError := &ServerError{}
		if err = json.Unmarshal(resp.Body, serverError); err != nil {
			return response, err
		}
		err = fmt.Errorf("code %d, message: %s: %w", response.StatusCode, serverError.Message, ErrVerifyPubKeyBadResponse)
		return response, err
	}

	// Decode the body of the response
	if err = json.Unmarshal(resp.Body, &response); err != nil {
		return response, err
	}

	// Invalid version?
	if len(response.BsvAlias) == 0 {
		err = ErrVerifyPubKeyMissingVersion
		return response, err
	}

	// Check basic requirements (alias@domain.tld)
	if response.Handle != alias+"@"+domain {
		err = fmt.Errorf("handle %s does not match %s: %w", response.Handle, alias+"@"+domain, ErrVerifyPubKeyHandleNotMatching)
		return response, err
	}

	// Check the PubKey length
	if len(response.PubKey) == 0 {
		err = ErrVerifyPubKeyMissingPubKeyValue
	} else if len(response.PubKey) != PubKeyLength {
		err = fmt.Errorf("length %d, expected: %d: %w", len(response.PubKey), PubKeyLength, ErrVerifyPubKeyInvalidLength)
	}

	return response, err
}
