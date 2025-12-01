package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bsv-blockchain/go-paymail/errors"

	"github.com/gin-gonic/gin"

	"github.com/bsv-blockchain/go-paymail"
)

type CallableCapability struct {
	Path    string
	Method  string
	Handler gin.HandlerFunc
}

type NestedCapabilitiesMap map[string]CallableCapabilitiesMap
type CallableCapabilitiesMap map[string]CallableCapability
type StaticCapabilitiesMap map[string]any

func (c *Configuration) SetGenericCapabilities() {
	_addCapabilities(c.callableCapabilities,
		CallableCapabilitiesMap{
			paymail.BRFCPaymentDestination: CallableCapability{
				Path:    fmt.Sprintf("/address/%s", PaymailAddressTemplate),
				Method:  http.MethodPost,
				Handler: c.resolveAddress,
			},
			paymail.BRFCPki: CallableCapability{
				Path:    fmt.Sprintf("/id/%s", PaymailAddressTemplate),
				Method:  http.MethodGet,
				Handler: c.showPKI,
			},
			paymail.BRFCPublicProfile: CallableCapability{
				Path:    fmt.Sprintf("/public-profile/%s", PaymailAddressTemplate),
				Method:  http.MethodGet,
				Handler: c.publicProfile,
			},
			paymail.BRFCVerifyPublicKeyOwner: CallableCapability{
				Path:    fmt.Sprintf("/verify-pubkey/%s/%s", PaymailAddressTemplate, PubKeyTemplate),
				Method:  http.MethodGet,
				Handler: c.verifyPubKey,
			},
		},
	)
	_addCapabilities(c.staticCapabilities,
		StaticCapabilitiesMap{
			paymail.BRFCSenderValidation: c.SenderValidationEnabled,
		},
	)
}

func (c *Configuration) SetP2PCapabilities() {
	_addCapabilities(c.callableCapabilities,
		CallableCapabilitiesMap{
			paymail.BRFCP2PTransactions: CallableCapability{
				Path:    fmt.Sprintf("/receive-transaction/%s", PaymailAddressTemplate),
				Method:  http.MethodPost,
				Handler: c.p2pReceiveTx,
			},
			paymail.BRFCP2PPaymentDestination: CallableCapability{
				Path:    fmt.Sprintf("/p2p-payment-destination/%s", PaymailAddressTemplate),
				Method:  http.MethodPost,
				Handler: c.p2pDestination,
			},
		},
	)
}

func (c *Configuration) SetBeefCapabilities() {
	_addCapabilities(c.callableCapabilities,
		CallableCapabilitiesMap{
			paymail.BRFCBeefTransaction: CallableCapability{
				Path:    fmt.Sprintf("/beef/%s", PaymailAddressTemplate),
				Method:  http.MethodPost,
				Handler: c.p2pReceiveBeefTx,
			},
		},
	)
}

func (c *Configuration) SetPikeContactCapabilities() {
	_addNestedCapabilities(c.nestedCapabilities,
		NestedCapabilitiesMap{
			paymail.BRFCPike: CallableCapabilitiesMap{
				paymail.BRFCPikeInvite: CallableCapability{
					Path:    fmt.Sprintf("/contact/invite/%s", PaymailAddressTemplate),
					Method:  http.MethodPost,
					Handler: c.pikeNewContact,
				},
			},
		},
	)
}

func (c *Configuration) SetPikePaymentCapabilities() {
	_addNestedCapabilities(c.nestedCapabilities,
		NestedCapabilitiesMap{
			paymail.BRFCPike: CallableCapabilitiesMap{
				paymail.BRFCPikeOutputs: CallableCapability{
					Path:    fmt.Sprintf("/pike/outputs/%s", PaymailAddressTemplate),
					Method:  http.MethodPost,
					Handler: c.pikeGetOutputTemplates,
				},
			},
		},
	)
}

func _addCapabilities[T any](base map[string]T, newCaps map[string]T) {
	for key, val := range newCaps {
		base[key] = val
	}
}

func _addNestedCapabilities(base NestedCapabilitiesMap, newCaps NestedCapabilitiesMap) {
	for key, val := range newCaps {
		if _, ok := base[key]; !ok {
			base[key] = make(CallableCapabilitiesMap)
		}
		_addCapabilities(base[key], val)
	}
}

// showCapabilities will return the service discovery results for the server
// and list all active capabilities of the Paymail server
//
// Specs: http://bsvalias.org/02-02-capability-discovery.html
func (c *Configuration) showCapabilities(context *gin.Context) {
	// Check the host (allowed, and used for capabilities response)
	// todo: bake this into middleware? This is protecting the "req" host name (like CORs)
	host := ""
	if context.Request.URL.IsAbs() || len(context.Request.URL.Host) == 0 {
		host = context.Request.Host
	} else {
		host = context.Request.URL.Host
	}

	if !c.IsAllowedDomain(host) {
		errors.ErrorResponse(context, errors.ErrDomainUnknown, c.Logger)
		return
	}

	capabilities, err := c.EnrichCapabilities(host)
	if err != nil {
		errors.ErrorResponse(context, err, c.Logger)
		return
	}

	context.JSON(http.StatusOK, capabilities)
}

// EnrichCapabilities will update the capabilities with the appropriate service url
func (c *Configuration) EnrichCapabilities(host string) (*paymail.CapabilitiesPayload, error) {
	serviceUrl, err := generateServiceURL(c.Prefix, host, c.APIVersion, c.ServiceName)
	if err != nil {
		return nil, err
	}
	payload := &paymail.CapabilitiesPayload{
		BsvAlias:     c.BSVAliasVersion,
		Capabilities: make(map[string]interface{}),
	}
	for key, cap := range c.staticCapabilities {
		payload.Capabilities[key] = cap
	}
	for key, cap := range c.callableCapabilities {
		payload.Capabilities[key] = serviceUrl + string(cap.Path)
	}
	for key, cap := range c.nestedCapabilities {
		payload.Capabilities[key] = make(map[string]interface{})
		for nestedKey, nestedCap := range cap {
			nestedObj, ok := payload.Capabilities[key].(map[string]interface{})
			if !ok {
				return nil, errors.ErrCastingNestedCapabilities
			}
			nestedObj[nestedKey] = serviceUrl + nestedCap.Path
		}
	}
	return payload, nil
}

func generateServiceURL(prefix, domain, apiVersion, serviceName string) (string, error) {
	if len(prefix) == 0 || len(domain) == 0 {
		return "", errors.ErrPrefixOrDomainMissing
	}
	strBuilder := new(strings.Builder)
	strBuilder.WriteString(prefix)
	strBuilder.WriteString(domain)
	if len(apiVersion) > 0 {
		strBuilder.WriteString("/")
		strBuilder.WriteString(apiVersion)
	}
	if len(serviceName) > 0 {
		strBuilder.WriteString("/")
		strBuilder.WriteString(serviceName)
	}

	return strBuilder.String(), nil
}
