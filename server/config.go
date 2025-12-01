package server

import (
	"slices"
	"strings"
	"time"

	"github.com/bsv-blockchain/go-paymail/errors"
	"github.com/rs/zerolog"

	"github.com/bsv-blockchain/go-paymail"
)

// Configuration paymail server configuration object
type Configuration struct {
	APIVersion                       string          `json:"api_version"`
	BasicRoutes                      *basicRoutes    `json:"basic_routes"`
	BSVAliasVersion                  string          `json:"bsv_alias_version"`
	PaymailDomains                   []*Domain       `json:"paymail_domains"`
	PaymailDomainsValidationDisabled bool            `json:"paymail_domains_validation_disabled"`
	Port                             int             `json:"port"`
	Prefix                           string          `json:"prefix"`
	SenderValidationEnabled          bool            `json:"sender_validation_enabled"`
	GenericCapabilitiesEnabled       bool            `json:"generic_capabilities_enabled"`
	P2PCapabilitiesEnabled           bool            `json:"p2p_capabilities_enabled"`
	BeefCapabilitiesEnabled          bool            `json:"beef_capabilities_enabled"`
	PikeContactCapabilitiesEnabled   bool            `json:"pike_contact_capabilities_enabled"`
	PikePaymentCapabilitiesEnabled   bool            `json:"pike_payment_capabilities_enabled"`
	ServiceName                      string          `json:"service_name"`
	Timeout                          time.Duration   `json:"timeout"`
	Logger                           *zerolog.Logger `json:"logger"`

	// private
	actions              PaymailServiceProvider
	pikeContactActions   PikeContactServiceProvider
	pikePaymentActions   PikePaymentServiceProvider
	nestedCapabilities   NestedCapabilitiesMap
	callableCapabilities CallableCapabilitiesMap
	staticCapabilities   StaticCapabilitiesMap
}

// Domain is the Paymail Domain information
type Domain struct {
	Name string `json:"name"`
}

// Validate will check that the configuration meets a minimum requirement to run the server
func (c *Configuration) Validate() error {

	// Requires domains for the server to run
	if len(c.PaymailDomains) == 0 && !c.PaymailDomainsValidationDisabled {
		return errors.ErrDomainMissing
	}

	// Requires a port
	if c.Port <= 0 {
		return errors.ErrPortMissing
	}

	// todo: validate the []domains

	// Sanitize and standardize the service name
	c.ServiceName = paymail.SanitizePathName(c.ServiceName)
	if len(c.ServiceName) == 0 {
		return errors.ErrServiceNameMissing
	}

	if c.BSVAliasVersion == "" {
		return errors.ErrBsvAliasMissing
	}

	if len(c.callableCapabilities) == 0 {
		return errors.ErrCapabilitiesMissing
	}

	return nil
}

// IsAllowedDomain will return true if it's an allowed paymail domain
func (c *Configuration) IsAllowedDomain(domain string) bool {
	if c.PaymailDomainsValidationDisabled {
		return true
	}

	var err error
	if domain, err = paymail.SanitizeDomain(domain); err != nil {
		c.Logger.Warn().Err(err).Msg("failed to sanitize domain")
		return false
	}

	return slices.ContainsFunc(c.PaymailDomains, func(d *Domain) bool {
		return strings.EqualFold(d.Name, domain)
	})
}

// AddDomain will add the domain if it does not exist
func (c *Configuration) AddDomain(domain string) (err error) {

	// Sanity check
	if len(domain) == 0 {
		return errors.ErrDomainMissing
	}

	// Sanitize and standardize
	domain, err = paymail.SanitizeDomain(domain)
	if err != nil {
		return
	}

	// Already exists?
	if c.IsAllowedDomain(domain) {
		return
	}

	// Create the domain
	c.PaymailDomains = append(c.PaymailDomains, &Domain{Name: domain})
	return
}

// NewConfig will make a new server configuration
// The serviceProvider must have registered necessary services before calling them (e.g., PikeServiceProvider has to be registered if Pike capabilities are supported)
func NewConfig(serviceProvider *PaymailServiceLocator, opts ...ConfigOps) (*Configuration, error) {

	// Check that a service provider is set
	if serviceProvider == nil {
		return nil, errors.ErrServiceProviderNil
	}

	// Create the base configuration
	config := defaultConfigOptions()

	// Overwrite defaults
	for _, opt := range opts {
		opt(config)
	}

	if config.GenericCapabilitiesEnabled {
		config.SetGenericCapabilities()
	}
	if config.P2PCapabilitiesEnabled {
		config.SetP2PCapabilities()
	}
	if config.BeefCapabilitiesEnabled {
		config.SetBeefCapabilities()
	}

	if config.PikeContactCapabilitiesEnabled {
		config.SetPikeContactCapabilities()
		config.pikeContactActions = serviceProvider.GetPikeContactService()
	}

	if config.PikePaymentCapabilitiesEnabled {
		config.SetPikePaymentCapabilities()
		config.pikePaymentActions = serviceProvider.GetPikePaymentService()
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Set the service provider
	config.actions = serviceProvider.GetPaymailService()

	config.Logger.Debug().Msg("New config loaded")
	return config, nil
}
