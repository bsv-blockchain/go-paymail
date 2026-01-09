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
	// ErrP2PInvalidURL is returned when P2P URL is invalid
	ErrP2PInvalidURL = errors.New("invalid url")
	// ErrP2PPaymentRequestNil is returned when payment request is nil
	ErrP2PPaymentRequestNil = errors.New("paymentRequest cannot be nil")
	// ErrP2PSatoshisRequired is returned when satoshis is required but not set
	ErrP2PSatoshisRequired = errors.New("satoshis is required")
	// ErrP2PMissingAlias is returned when alias is missing
	ErrP2PMissingAlias = errors.New("missing alias")
	// ErrP2PMissingDomain is returned when domain is missing
	ErrP2PMissingDomain = errors.New("missing domain")
	// ErrP2PAddressNotFound is returned when paymail address is not found
	ErrP2PAddressNotFound = errors.New("paymail address not found")
	// ErrP2PBadResponse is returned when receiving bad response from paymail provider
	ErrP2PBadResponse = errors.New("bad response from paymail provider")
	// ErrP2PMissingReference is returned when reference is missing from response
	ErrP2PMissingReference = errors.New("missing a returned reference value")
	// ErrP2PMissingOutput is returned when output is missing from response
	ErrP2PMissingOutput = errors.New("missing a returned output")
	// ErrP2PMissingScript is returned when script is missing from output
	ErrP2PMissingScript = errors.New("script was missing from output")
	// ErrP2PInvalidOutputScript is returned when output script is invalid
	ErrP2PInvalidOutputScript = errors.New("invalid output script, missing an address")
)

/*
Example:
{
  "satoshis": 1000100
}
*/

// PaymentRequest is the request body for the P2P payment request
type PaymentRequest struct {
	Satoshis uint64 `json:"satoshis"` // The amount, in Satoshis, that the sender intends to transfer to the receiver
}

// PaymentDestinationResponse is the response from the GetP2PPaymentDestination() request
//
// The reference is unique for the payment destination request
type PaymentDestinationResponse struct {
	StandardResponse
	PaymentDestinationPayload
}

// PaymentDestinationPayload is the payload from the response
//
// The reference is unique for the payment destination request
type PaymentDestinationPayload struct {
	Outputs   []*PaymentOutput `json:"outputs"`   // A list of outputs
	Reference string           `json:"reference"` // A reference for the payment, created by the receiver of the transaction
}

// PaymentOutput is returned inside the payment destination response
//
// There can be several outputs in one response based on the amount of satoshis being transferred and
// the rules in place by the Paymail provider
type PaymentOutput struct {
	Address  string `json:"address,omitempty"`  // Hex encoded locking script
	Satoshis uint64 `json:"satoshis,omitempty"` // Number of satoshis for that output
	Script   string `json:"script"`             // Hex encoded locking script
}

// GetP2PPaymentDestination will return list of outputs for the P2P transactions to use
//
// Specs: https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html
func (c *Client) GetP2PPaymentDestination(p2pURL, alias, domain string,
	paymentRequest *PaymentRequest,
) (response *PaymentDestinationResponse, err error) {
	// Require a valid url
	if len(p2pURL) == 0 || !strings.Contains(p2pURL, "https://") {
		err = fmt.Errorf("%s: %s: %w", "invalid url", p2pURL, ErrP2PInvalidURL)
		return response, err
	}

	// Basic requirements for request
	if paymentRequest == nil {
		err = ErrP2PPaymentRequestNil
		return response, err
	} else if paymentRequest.Satoshis == 0 {
		err = ErrP2PSatoshisRequired
		return response, err
	} else if len(alias) == 0 {
		err = ErrP2PMissingAlias
		return response, err
	} else if len(domain) == 0 {
		err = ErrP2PMissingDomain
		return response, err
	}

	// Set the base url and path, assuming the url is from the prior GetCapabilities() request
	// https://<host-discovery-target>/api/rawtx/{alias}@{domain.tld}
	// https://<host-discovery-target>/api/p2p-payment-destination/{alias}@{domain.tld}
	reqURL := replaceAliasDomain(p2pURL, alias, domain)

	// Fire the POST request
	var resp StandardResponse
	if resp, err = c.postRequest(reqURL, paymentRequest); err != nil {
		return response, err
	}

	// Start the response
	response = &PaymentDestinationResponse{StandardResponse: resp}

	// Test the status code
	if response.StatusCode != http.StatusOK &&
		response.StatusCode != http.StatusNotModified {

		// Paymail address not found?
		if response.StatusCode == http.StatusNotFound {
			err = ErrP2PAddressNotFound
		} else {
			serverError := &ServerError{}
			if err = json.Unmarshal(resp.Body, serverError); err != nil {
				return response, err
			}
			err = fmt.Errorf(
				"code %d, message: %s: %w",
				response.StatusCode, serverError.Message, ErrP2PBadResponse,
			)
		}

		return response, err
	}

	// Decode the body of the response
	if err = json.Unmarshal(resp.Body, &response); err != nil {
		return response, err
	}

	// Check for a reference number
	if len(response.Reference) == 0 {
		err = ErrP2PMissingReference
		return response, err
	}

	// No outputs?
	if len(response.Outputs) == 0 {
		err = ErrP2PMissingOutput
		return response, err
	}

	// Loop all outputs
	for index, out := range response.Outputs {
		// No script returned
		if len(out.Script) == 0 {
			err = fmt.Errorf("output %d: %w", index, ErrP2PMissingScript)
			return response, err
		}

		var sc *script.Script
		sc, err = script.NewFromHex(out.Script)
		if err != nil {
			return response, err
		}

		var addresses []string
		addresses, err = sc.Addresses()
		if err != nil || len(addresses) == 0 {
			err = ErrP2PInvalidOutputScript
			return response, err
		}

		response.Outputs[index].Address = addresses[0]
	}

	return response, err
}
