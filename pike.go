package paymail

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
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
			return nil, errors.New("paymail address not found")
		} else {
			return nil, c.prepareServerErrorResponse(&response)
		}
	}

	return &PikeContactRequestResponse{response}, nil
}

func (c *Client) validateUrlWithPaymail(url, alias, domain string) error {
	if len(url) == 0 || !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("invalid url: %s", url)
	} else if alias == "" {
		return errors.New("missing alias")
	} else if domain == "" {
		return errors.New("missing domain")
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

	return fmt.Errorf("bad response from paymail provider: code %d, %s", response.StatusCode, details)
}

func (r *PikeContactRequestPayload) validate() error {
	if r.FullName == "" {
		return errors.New("missing full name")
	}
	if r.Paymail == "" {
		return errors.New("missing paymail address")
	}

	return ValidatePaymail(r.Paymail)
}

// GetOutputsTemplate calls the PIKE capability outputs subcapability
func (c *Client) GetOutputsTemplate(pikeURL, alias, domain string, payload *PikePaymentOutputsPayload) (response *PikePaymentOutputsResponse, err error) {
	// Require a valid URL
	if len(pikeURL) == 0 || !strings.Contains(pikeURL, "https://") {
		err = fmt.Errorf("invalid url: %s", pikeURL)
		return response, err
	}

	// Basic requirements for request
	if payload == nil {
		err = errors.New("payload cannot be nil")
		return response, err
	} else if payload.Amount == 0 {
		err = errors.New("amount is required")
		return response, err
	} else if len(alias) == 0 {
		err = errors.New("missing alias")
		return response, err
	} else if len(domain) == 0 {
		err = errors.New("missing domain")
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
		return nil, fmt.Errorf("bad response from PIKE outputs: code %d", resp.StatusCode)
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
