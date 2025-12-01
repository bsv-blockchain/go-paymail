package paymail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

/*
Default:

{
  "bsvalias": "1.0",
  "capabilities": {
	"6745385c3fc0": false,
	"pki": "https://bsvalias.example.org/{alias}@{domain.tld}/id",
	"paymentDestination": "https://bsvalias.example.org/{alias}@{domain.tld}/payment-destination"
  }
}
*/

// CapabilitiesResponse is the full response returned
type CapabilitiesResponse struct {
	StandardResponse
	CapabilitiesPayload
}

// CapabilitiesPayload is the actual payload response
type CapabilitiesPayload struct {
	BsvAlias     string                 `json:"bsvalias"`     // Version of the bsvalias
	Capabilities map[string]interface{} `json:"capabilities"` // Raw list of the capabilities
	Pike         *PikeCapability        `json:"pike,omitempty"`
}

// PikeCapability represents the structure of the PIKE capability
type PikeCapability struct {
	Invite  *string `json:"invite,omitempty"`
	Outputs *string `json:"outputs,omitempty"`
}

// PikeOutputs represents the structure of the PIKE outputs
type PikeOutputs struct {
	URL string `json:"url"`
}

// Has will check if a BRFC ID (or alternate) is found in the list of capabilities
//
// Alternate is used for example: "pki" is also BRFC "0c4339ef99c2"
func (c *CapabilitiesPayload) Has(brfcID, alternateID string) bool {
	for key := range c.Capabilities {
		if key == brfcID || (len(alternateID) > 0 && key == alternateID) {
			return true
		}
	}
	return false
}

// getValue will return the value (if found) from the capability (url or bool)
//
// Alternate is used for IE: pki (it breaks convention of using the BRFC ID)
func (c *CapabilitiesPayload) getValue(brfcID, alternateID string) (bool, interface{}) {
	for key, val := range c.Capabilities {
		if key == brfcID || (len(alternateID) > 0 && key == alternateID) {
			return true, val
		}
	}
	return false, nil
}

// GetString will perform getValue() but cast to a string if found
//
// Returns an empty string if not found
func (c *CapabilitiesPayload) GetString(brfcID, alternateID string) string {
	if ok, val := c.getValue(brfcID, alternateID); ok {
		return val.(string)
	}
	return ""
}

// GetBool will perform getValue() but cast to a bool if found
//
// Returns false if not found
func (c *CapabilitiesPayload) GetBool(brfcID, alternateID string) bool {
	if ok, val := c.getValue(brfcID, alternateID); ok {
		return val.(bool)
	}
	return false
}

// GetCapabilities will return a list of capabilities for a given domain & port
//
// Specs: http://bsvalias.org/02-02-capability-discovery.html
func (c *Client) GetCapabilities(target string, port int) (response *CapabilitiesResponse, err error) {
	// Basic requirements for the request
	if len(target) == 0 {
		err = fmt.Errorf("missing target")
		return response, err
	} else if port == 0 {
		err = fmt.Errorf("missing port")
		return response, err
	}

	// Set the base url and path
	// https://<host-discovery-target>:<host-discovery-port>/.well-known/bsvalias[network]
	reqURL := fmt.Sprintf("https://%s:%d/.well-known/%s%s", target, port, DefaultServiceName, c.options.network.URLSuffix())

	// Fire the GET request
	var resp StandardResponse
	if resp, err = c.getRequest(reqURL); err != nil {
		return response, err
	}

	// Start the response
	response = &CapabilitiesResponse{StandardResponse: resp}

	// Test the status code (200 or 304 is valid)
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotModified {
		serverError := &ServerError{}
		if err = json.Unmarshal(resp.Body, serverError); err != nil {
			return response, err
		}
		err = fmt.Errorf("bad response from paymail provider: code %d, message: %s", response.StatusCode, serverError.Message)
		return response, err
	}

	// Decode the body of the response
	if err = json.Unmarshal(resp.Body, &response); err != nil {

		// Invalid character (sometimes quote related: U+0022 vs U+201C)
		if strings.Contains(err.Error(), "invalid character") {

			// Replace any invalid quotes
			bodyString := strings.Replace(strings.Replace(string(resp.Body), `“`, `"`, -1), `”`, `"`, -1)

			// Parse again after fixing quotes
			if err = json.Unmarshal([]byte(bodyString), &response); err != nil {
				return response, err
			}
		}

		// Still have an error?
		if err != nil {
			return response, err
		}
	}

	// Invalid version detected
	if len(response.BsvAlias) == 0 {
		err = fmt.Errorf("missing %s version", DefaultServiceName)
		return response, err
	}

	// Parse PIKE capability
	if err = parsePikeCapability(response); err != nil {
		return response, err
	}

	return response, err
}

// ExtractPikeOutputsURL extracts the outputs URL from the PIKE capability
func (c *CapabilitiesPayload) ExtractPikeOutputsURL() string {
	if c.Pike != nil {
		return *c.Pike.Outputs
	}
	return ""
}

// ExtractPikeInviteURL extracts the invite URL from the PIKE capability
func (c *CapabilitiesPayload) ExtractPikeInviteURL() string {
	if c.Pike != nil {
		return *c.Pike.Invite
	}
	return ""
}

// parsePikeCapability parses the PIKE capability from the capabilities response
func parsePikeCapability(response *CapabilitiesResponse) error {
	if pike, ok := response.Capabilities[BRFCPike].(map[string]interface{}); ok {
		response.Pike = &PikeCapability{}

		if inviteStr, ok := pike["invite"].(string); ok {
			response.Pike.Invite = &inviteStr
		}

		// TODO: Ensure the outputs field is required when the 'PIKE' transaction with linked keys is implemented.
		if outputsStr, ok := pike["outputs"].(string); ok {
			response.Pike.Outputs = &outputsStr
		}
	}
	return nil
}
