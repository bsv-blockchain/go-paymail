package paymail

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	// ErrPikeAddressNotFound is returned when paymail address is not found
	ErrPikeAddressNotFound = errors.New("paymail address not found")
	// ErrPikeMissingAlias is returned when alias is missing
	ErrPikeMissingAlias = errors.New("missing alias")
	// ErrPikeMissingDomain is returned when domain is missing
	ErrPikeMissingDomain = errors.New("missing domain")
	// ErrPikeMissingFullName is returned when full name is missing
	ErrPikeMissingFullName = errors.New("missing full name")
	// ErrPikeMissingPaymail is returned when paymail address is missing
	ErrPikeMissingPaymail = errors.New("missing paymail address")
	// ErrPikePayloadNil is returned when payload is nil
	ErrPikePayloadNil = errors.New("payload cannot be nil")
	// ErrPikeAmountRequired is returned when amount is required but not set
	ErrPikeAmountRequired = errors.New("amount is required")
	// ErrPikeInvalidURL is returned when URL is invalid
	ErrPikeInvalidURL = errors.New("invalid url")
	// ErrPikeBadResponse is returned when paymail provider returns bad response
	ErrPikeBadResponse = errors.New("bad response from paymail provider")
	// ErrPikeBadOutputsResponse is returned when PIKE outputs returns bad response
	ErrPikeBadOutputsResponse = errors.New("bad response from PIKE outputs")
)

// PikeContactRequestResponse is PIKE wrapper for StandardResponse
type PikeContactRequestResponse struct {
	StandardResponse
}

// PikeContactRequestPayload is a payload used to request a contact
type PikeContactRequestPayload struct {
	FullName string `json:"fullName"`
	Paymail  string `json:"paymail"`
}

// PikePaymentOutputsPayload is a payload needed to get payment outputs
type PikePaymentOutputsPayload struct {
	SenderPaymail string `json:"senderPaymail"`
	Amount        uint64 `json:"amount"`
}

// PikePaymentOutputsResponse is a response which contain output templates
type PikePaymentOutputsResponse struct {
	Outputs   []*OutputTemplate `json:"outputs"`
	Reference string            `json:"reference"`
}

// OutputTemplate is a single output template with satoshis
type OutputTemplate struct {
	Script   string `json:"script"`
	Satoshis uint64 `json:"satoshis"`
}

func (c *Client) AddContactRequest(url, alias, domain string, request *PikeContactRequestPayload) (*PikeContactRequestResponse, error) {
	if err := c.validateUrlWithPaymail(url, alias, domain); err != nil {
		return nil, err
	}

	if err := request.validate(); err != nil {
		return nil, err
	}

	// Set the base url and path, assuming the url is from the prior GetCapabilities() request
	// https://<host-discovery-target>/{alias}@{domain.tld}/id
	reqURL := replaceAliasDomain(url, alias, domain)

	response, err := c.postRequest(reqURL, request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		if response.StatusCode == http.StatusNotFound {
			return nil, ErrPikeAddressNotFound
		} else {
			return nil, c.prepareServerErrorResponse(&response)
		}
	}

	return &PikeContactRequestResponse{response}, nil
}

func (c *Client) validateUrlWithPaymail(url, alias, domain string) error {
	if len(url) == 0 || !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("url %s: %w", url, ErrPikeInvalidURL)
	} else if alias == "" {
		return ErrPikeMissingAlias
	} else if domain == "" {
		return ErrPikeMissingDomain
	}
	return nil
}

func (c *Client) prepareServerErrorResponse(response *StandardResponse) error {
	var details string

	serverError := &ServerError{}
	if err := json.Unmarshal(response.Body, serverError); err != nil || serverError.Message == "" {
		details = fmt.Sprintf("body: %s", string(response.Body))
	} else {
		details = fmt.Sprintf("message: %s", serverError.Message)
	}

	return fmt.Errorf("code %d, %s: %w", response.StatusCode, details, ErrPikeBadResponse)
}

func (r *PikeContactRequestPayload) validate() error {
	if r.FullName == "" {
		return ErrPikeMissingFullName
	}
	if r.Paymail == "" {
		return ErrPikeMissingPaymail
	}

	return ValidatePaymail(r.Paymail)
}

// GetOutputsTemplate calls the PIKE capability outputs subcapability
func (c *Client) GetOutputsTemplate(pikeURL, alias, domain string, payload *PikePaymentOutputsPayload) (response *PikePaymentOutputsResponse, err error) {
	// Require a valid URL
	if len(pikeURL) == 0 || !strings.Contains(pikeURL, "https://") {
		err = fmt.Errorf("url %s: %w", pikeURL, ErrPikeInvalidURL)
		return response, err
	}

	// Basic requirements for request
	if payload == nil {
		err = ErrPikePayloadNil
		return response, err
	} else if payload.Amount == 0 {
		err = ErrPikeAmountRequired
		return response, err
	} else if len(alias) == 0 {
		err = ErrPikeMissingAlias
		return response, err
	} else if len(domain) == 0 {
		err = ErrPikeMissingDomain
		return response, err
	}

	// Set the base URL and path, assuming the URL is from the prior GetCapabilities() request
	reqURL := replaceAliasDomain(pikeURL, alias, domain)

	// Fire the POST request
	var resp StandardResponse
	if resp, err = c.postRequest(reqURL, payload); err != nil {
		return response, err
	}

	// Test the status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("code %d: %w", resp.StatusCode, ErrPikeBadOutputsResponse)
	}

	// Decode the body of the response
	outputs := &PikePaymentOutputsResponse{}
	if err = json.Unmarshal(resp.Body, outputs); err != nil {
		return nil, err
	}

	return outputs, nil
}

// AddInviteRequest sends a contact request using the invite URL from capabilities
func (c *Client) AddInviteRequest(inviteURL, alias, domain string, request *PikeContactRequestPayload) (*PikeContactRequestResponse, error) {
	return c.AddContactRequest(inviteURL, alias, domain, request)
}
