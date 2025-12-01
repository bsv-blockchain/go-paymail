package server

import (
	"time"

	"github.com/rs/zerolog"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/bsv-blockchain/go-paymail/logging"
)

// ConfigOps allow functional options to be supplied
// that overwrite default options.
type ConfigOps func(c *Configuration)

// defaultConfigOptions will return a Configuration struct with the default settings
//
// Useful for starting with the default and then modifying as needed
func defaultConfigOptions() *Configuration {
	return &Configuration{
		APIVersion:                       DefaultAPIVersion,
		BasicRoutes:                      &basicRoutes{},
		BSVAliasVersion:                  paymail.DefaultBsvAliasVersion,
		PaymailDomainsValidationDisabled: false,
		Port:                             DefaultServerPort,
		Prefix:                           DefaultPrefix,
		SenderValidationEnabled:          DefaultSenderValidation,
		GenericCapabilitiesEnabled:       true,
		P2PCapabilitiesEnabled:           false,
		BeefCapabilitiesEnabled:          false,
		PikeContactCapabilitiesEnabled:   false,
		PikePaymentCapabilitiesEnabled:   false,
		ServiceName:                      paymail.DefaultServiceName,
		Timeout:                          DefaultTimeout,
		Logger:                           logging.GetDefaultLogger(),
		nestedCapabilities:               make(NestedCapabilitiesMap),
		callableCapabilities:             make(CallableCapabilitiesMap),
		staticCapabilities:               make(StaticCapabilitiesMap),
	}
}

// WithGenericCapabilities will load the generic Paymail capabilities
func WithGenericCapabilities() ConfigOps {
	return func(c *Configuration) {
		c.GenericCapabilitiesEnabled = true
	}
}

// WithP2PCapabilities will load the generic & p2p capabilities
func WithP2PCapabilities() ConfigOps {
	return func(c *Configuration) {
		c.GenericCapabilitiesEnabled = true
		c.P2PCapabilitiesEnabled = true
	}
}

// WithBeefCapabilities will load the beef capabilities
func WithBeefCapabilities() ConfigOps {
	return func(c *Configuration) {
		c.BeefCapabilitiesEnabled = true
	}
}

// WithPikeContactCapabilities will load the PIKE capabilities
func WithPikeContactCapabilities() ConfigOps {
	return func(c *Configuration) {
		c.PikeContactCapabilitiesEnabled = true
	}
}

// WithPikePaymentCapabilities will load the PIKE capabilities
func WithPikePaymentCapabilities() ConfigOps {
	return func(c *Configuration) {
		c.PikePaymentCapabilitiesEnabled = true
	}
}

// WithCapabilities will modify the capabilities
func WithCapabilities(customCapabilities map[string]any) ConfigOps {
	return func(c *Configuration) {
		for key, cap := range customCapabilities {
			switch typedCap := cap.(type) {
			case CallableCapability:
				c.callableCapabilities[key] = typedCap
			default:
				c.staticCapabilities[key] = typedCap
			}
		}
	}
}

// WithBasicRoutes will turn on all the basic routes
func WithBasicRoutes() ConfigOps {
	return func(c *Configuration) {
		c.BasicRoutes = &basicRoutes{
			Add404Route:    true,
			AddHealthRoute: true,
			AddIndexRoute:  true,
			AddNotAllowed:  true,
		}
	}
}

// WithTimeout will set a custom timeout
func WithTimeout(timeout time.Duration) ConfigOps {
	return func(c *Configuration) {
		if timeout > 0 {
			c.Timeout = timeout
		}
	}
}

// WithServiceName will set a custom service name
func WithServiceName(serviceName string) ConfigOps {
	return func(c *Configuration) {
		if len(serviceName) > 0 {
			c.ServiceName = serviceName
		}
	}
}

// WithSenderValidation will enable sender validation
func WithSenderValidation() ConfigOps {
	return func(c *Configuration) {
		c.SenderValidationEnabled = true
	}
}

// WithDomain will add the domain if not found
func WithDomain(domain string) ConfigOps {
	return func(c *Configuration) {
		if len(domain) > 0 {
			// todo: attempt to add, but cannot return the error
			_ = c.AddDomain(domain)
		}
	}
}

// WithPort will overwrite the default port
func WithPort(port int) ConfigOps {
	return func(c *Configuration) {
		if port > 0 {
			c.Port = port
		}
	}
}

// WithDomainValidationDisabled will disable checking domains (from request for allowed domains)
func WithDomainValidationDisabled() ConfigOps {
	return func(c *Configuration) {
		c.PaymailDomainsValidationDisabled = true
	}
}

// WithLogger will set a custom logger
func WithLogger(logger *zerolog.Logger) ConfigOps {
	return func(c *Configuration) {
		c.Logger = logger
	}
}
