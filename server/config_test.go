package server

import (
	"testing"
	"time"

	"github.com/bsv-blockchain/go-paymail/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bsv-blockchain/go-paymail"
)

// testConfig loads a basic test configuration
func testConfig(t *testing.T, domain string) *Configuration {
	sl := PaymailServiceLocator{}
	sl.RegisterPaymailService(new(mockServiceProvider))
	sl.RegisterPikeContactService(new(mockServiceProvider))
	sl.RegisterPikePaymentService(new(mockServiceProvider))

	c, err := NewConfig(
		&sl,
		WithDomain(domain),
	)
	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}

// TestConfiguration_Validate will test the method Validate()
func TestConfiguration_Validate(t *testing.T) {
	t.Parallel()

	t.Run("missing domain", func(t *testing.T) {
		c := &Configuration{}
		err := c.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrDomainMissing)
	})

	t.Run("missing port", func(t *testing.T) {
		c := &Configuration{
			PaymailDomains: []*Domain{{Name: "test.com"}},
		}
		err := c.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrPortMissing)
	})

	t.Run("missing service name", func(t *testing.T) {
		c := &Configuration{
			Port:           12345,
			PaymailDomains: []*Domain{{Name: "test.com"}},
		}
		err := c.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrServiceNameMissing)
	})

	t.Run("invalid service name", func(t *testing.T) {
		c := &Configuration{
			Port:           12345,
			ServiceName:    "$*%*",
			PaymailDomains: []*Domain{{Name: "test.com"}},
		}
		err := c.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrServiceNameMissing)
	})

	t.Run("missing bsv alias", func(t *testing.T) {
		c := &Configuration{
			Port:           12345,
			ServiceName:    "test",
			PaymailDomains: []*Domain{{Name: "test.com"}},
		}
		err := c.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrBsvAliasMissing)
	})

	t.Run("missing capabilities", func(t *testing.T) {
		c := &Configuration{
			Port:                 12345,
			ServiceName:          "test",
			PaymailDomains:       []*Domain{{Name: "test.com"}},
			BSVAliasVersion:      paymail.DefaultBsvAliasVersion,
			callableCapabilities: nil,
			staticCapabilities:   nil,
		}
		err := c.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrCapabilitiesMissing)
	})

	t.Run("zero capabilities", func(t *testing.T) {
		c := &Configuration{
			Port:                 12345,
			ServiceName:          "test",
			PaymailDomains:       []*Domain{{Name: "test.com"}},
			BSVAliasVersion:      paymail.DefaultBsvAliasVersion,
			callableCapabilities: make(CallableCapabilitiesMap),
			staticCapabilities:   make(StaticCapabilitiesMap),
		}
		err := c.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrCapabilitiesMissing)
	})

	t.Run("basic valid configuration", func(t *testing.T) {
		c := &Configuration{
			Port:                 12345,
			ServiceName:          "test",
			BSVAliasVersion:      paymail.DefaultBsvAliasVersion,
			PaymailDomains:       []*Domain{{Name: "test.com"}},
			callableCapabilities: make(CallableCapabilitiesMap),
			staticCapabilities:   make(StaticCapabilitiesMap),
		}
		c.SetGenericCapabilities()
		err := c.Validate()
		require.NoError(t, err)
	})

	t.Run("configuration with domain validation disabled", func(t *testing.T) {
		c := &Configuration{
			Port:                             12345,
			ServiceName:                      "test",
			BSVAliasVersion:                  paymail.DefaultBsvAliasVersion,
			PaymailDomains:                   []*Domain{},
			PaymailDomainsValidationDisabled: false,
			callableCapabilities:             make(CallableCapabilitiesMap),
			staticCapabilities:               make(StaticCapabilitiesMap),
		}
		c.SetGenericCapabilities()
		assert.False(t, c.PaymailDomainsValidationDisabled)
		err := c.Validate()
		assert.ErrorIs(t, err, errors.ErrDomainMissing)

		c.PaymailDomainsValidationDisabled = true
		err = c.Validate()
		assert.NoError(t, err)
	})

	t.Run("configuration with SenderValidationEnabled", func(t *testing.T) {
		c := &Configuration{
			Port:                    12345,
			Prefix:                  "https://",
			ServiceName:             "test",
			BSVAliasVersion:         paymail.DefaultBsvAliasVersion,
			PaymailDomains:          []*Domain{{Name: "test.com"}},
			SenderValidationEnabled: false,
			callableCapabilities:    make(CallableCapabilitiesMap),
			staticCapabilities:      make(StaticCapabilitiesMap),
		}
		c.SetGenericCapabilities()
		err := c.Validate()
		assert.NoError(t, err)
		caps, err := c.EnrichCapabilities("test.com")
		assert.NoError(t, err)
		assert.False(t, caps.Capabilities[paymail.BRFCSenderValidation].(bool))

		c.SenderValidationEnabled = true
		c.SetGenericCapabilities()
		err = c.Validate()
		assert.NoError(t, err)
		caps, err = c.EnrichCapabilities("test.com")
		assert.NoError(t, err)
		assert.True(t, caps.Capabilities[paymail.BRFCSenderValidation].(bool))
	})
}

// TestConfiguration_IsAllowedDomain will test the method IsAllowedDomain()
func TestConfiguration_IsAllowedDomain(t *testing.T) {
	t.Parallel()

	t.Run("empty domain", func(t *testing.T) {
		c := testConfig(t, "test.com")
		require.NotNil(t, c)

		success := c.IsAllowedDomain("")
		assert.Equal(t, false, success)
	})

	t.Run("domain found", func(t *testing.T) {
		c := testConfig(t, "test.com")
		require.NotNil(t, c)

		success := c.IsAllowedDomain("test.com")
		assert.Equal(t, true, success)
	})

	t.Run("sanitized domain found", func(t *testing.T) {
		c := testConfig(t, "test.com")
		require.NotNil(t, c)

		success := c.IsAllowedDomain("WWW.test.COM")
		assert.Equal(t, true, success)
	})

	t.Run("both domains are sanitized", func(t *testing.T) {
		c := testConfig(t, "WwW.Test.Com")
		require.NotNil(t, c)

		success := c.IsAllowedDomain("WWW.test.COM")
		assert.Equal(t, true, success)
	})

	t.Run("domain validation on", func(t *testing.T) {
		c := testConfig(t, "WwW.Test.Com")
		c.PaymailDomainsValidationDisabled = false
		require.NotNil(t, c)

		assert.Equal(t, true, c.IsAllowedDomain("test.com"))
		assert.Equal(t, false, c.IsAllowedDomain("test2.com"))
	})

	t.Run("domain validation off", func(t *testing.T) {
		c := testConfig(t, "WwW.Test.Com")
		c.PaymailDomainsValidationDisabled = true
		require.NotNil(t, c)

		assert.Equal(t, true, c.IsAllowedDomain("test.com"))
		assert.Equal(t, true, c.IsAllowedDomain("test2.com"))
	})
}

// TestConfiguration_AddDomain will test the method AddDomain()
func TestConfiguration_AddDomain(t *testing.T) {
	t.Parallel()

	t.Run("no domain", func(t *testing.T) {
		testDomain := "test.com"
		c := testConfig(t, testDomain)
		require.NotNil(t, c)

		err := c.AddDomain("")
		assert.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrDomainMissing)
	})

	t.Run("sanitized domain", func(t *testing.T) {
		testDomain := "WWW.TEST.COM"
		addDomain := "testER.com"
		c := testConfig(t, testDomain)
		require.NotNil(t, c)

		err := c.AddDomain(addDomain)
		require.NoError(t, err)

		assert.Equal(t, 2, len(c.PaymailDomains))
		assert.Equal(t, "test.com", c.PaymailDomains[0].Name)
		assert.Equal(t, "tester.com", c.PaymailDomains[1].Name)
	})

	t.Run("domain already exists", func(t *testing.T) {
		testDomain := "test.com"
		addDomain := "test.com"
		c := testConfig(t, testDomain)
		require.NotNil(t, c)

		err := c.AddDomain(addDomain)
		require.NoError(t, err)

		assert.Equal(t, 1, len(c.PaymailDomains))
		assert.Equal(t, "test.com", c.PaymailDomains[0].Name)
	})
}

// TestConfiguration_EnrichCapabilities will test the method EnrichCapabilities()
func TestConfiguration_EnrichCapabilities(t *testing.T) {
	t.Parallel()

	t.Run("basic enrich", func(t *testing.T) {
		testDomain := "test.com"
		c := testConfig(t, testDomain)
		require.NotNil(t, c)

		caps, err := c.EnrichCapabilities(testDomain)
		assert.NoError(t, err)
		assert.Equal(t, 5, len(caps.Capabilities))
		assert.Equal(t, "https://"+testDomain+"/v1/bsvalias/address/{alias}@{domain.tld}", caps.Capabilities[paymail.BRFCPaymentDestination])
		assert.Equal(t, "https://"+testDomain+"/v1/bsvalias/id/{alias}@{domain.tld}", caps.Capabilities[paymail.BRFCPki])
		assert.Equal(t, "https://"+testDomain+"/v1/bsvalias/public-profile/{alias}@{domain.tld}", caps.Capabilities[paymail.BRFCPublicProfile])
		assert.Equal(t, "https://"+testDomain+"/v1/bsvalias/verify-pubkey/{alias}@{domain.tld}/{pubkey}", caps.Capabilities[paymail.BRFCVerifyPublicKeyOwner])
		assert.Equal(t, false, caps.Capabilities[paymail.BRFCSenderValidation])
	})

	t.Run("multiple times", func(t *testing.T) {
		testDomain := "test.com"
		c := testConfig(t, testDomain)
		require.NotNil(t, c)

		caps, err := c.EnrichCapabilities(testDomain)
		assert.NoError(t, err)
		assert.Equal(t, 5, len(caps.Capabilities))

		caps, err = c.EnrichCapabilities(testDomain)
		assert.NoError(t, err)
		assert.Equal(t, 5, len(caps.Capabilities))
	})

	t.Run("empty domain and prefix", func(t *testing.T) {
		testDomain := "test.com"
		c := testConfig(t, testDomain)
		require.NotNil(t, c)

		c.Prefix = ""
		_, err := c.EnrichCapabilities("")
		assert.Error(t, err)
	})
}

// TestNewConfig will test the method NewConfig()
func TestNewConfig(t *testing.T) {
	t.Parallel()

	t.Run("no values and no provider", func(t *testing.T) {
		c, err := NewConfig(nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrServiceProviderNil)
		assert.Nil(t, c)
	})

	t.Run("missing domain", func(t *testing.T) {
		c, err := NewConfig(&PaymailServiceLocator{})
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrDomainMissing)
		assert.Nil(t, c)
	})

	t.Run("valid client - minimum options", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))
		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 4, len(c.callableCapabilities))
		assert.Equal(t, "test.com", c.PaymailDomains[0].Name)
	})

	t.Run("custom port", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))
		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithPort(12345),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 12345, c.Port)
	})

	t.Run("custom timeout", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))
		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithTimeout(10*time.Second),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 10*time.Second, c.Timeout)
	})

	t.Run("custom service name", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))
		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithServiceName("custom"),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, "custom", c.ServiceName)
	})

	t.Run("sender validation enabled", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))
		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithSenderValidation(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, true, c.SenderValidationEnabled)
	})

	t.Run("with p2p capabilities", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))
		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithP2PCapabilities(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 6, len(c.callableCapabilities))
	})

	t.Run("with custom capabilities", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))

		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithCapabilities(map[string]any{
				"test": true,
				"callable": CallableCapability{
					Path:    "/test",
					Method:  "GET",
					Handler: nil,
				},
			}),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 5, len(c.callableCapabilities))
		assert.Equal(t, 2, len(c.staticCapabilities))
		assert.True(t, c.staticCapabilities["test"].(bool))
		assert.Equal(t, "/test", c.callableCapabilities["callable"].Path)
	})

	t.Run("with beef capabilities", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))

		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithP2PCapabilities(),
			WithBeefCapabilities(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 7, len(c.callableCapabilities))
	})

	t.Run("with basic routes", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))

		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithBasicRoutes(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		require.NotNil(t, c.BasicRoutes)
		assert.Equal(t, true, c.BasicRoutes.Add404Route)
		assert.Equal(t, true, c.BasicRoutes.AddIndexRoute)
		assert.Equal(t, true, c.BasicRoutes.AddHealthRoute)
		assert.Equal(t, true, c.BasicRoutes.AddNotAllowed)
	})

	t.Run("domain validation disabled", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))

		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithPort(12345),
			WithDomainValidationDisabled(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 12345, c.Port)
		assert.Equal(t, true, c.PaymailDomainsValidationDisabled)
	})

	t.Run("with pike contact capabilities", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))
		sl.RegisterPikeContactService(new(mockServiceProvider))

		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithP2PCapabilities(),
			WithPikeContactCapabilities(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 6, len(c.callableCapabilities))
		assert.Equal(t, 1, len(c.nestedCapabilities))
	})

	t.Run("with pike payment capabilities", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))
		sl.RegisterPikePaymentService(new(mockServiceProvider))

		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithP2PCapabilities(),
			WithPikePaymentCapabilities(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 6, len(c.callableCapabilities))
		assert.Equal(t, 1, len(c.nestedCapabilities))
	})

	t.Run("with both pike capabilities", func(t *testing.T) {
		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))
		sl.RegisterPikeContactService(new(mockServiceProvider))
		sl.RegisterPikePaymentService(new(mockServiceProvider))

		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithP2PCapabilities(),
			WithPikeContactCapabilities(),
			WithPikePaymentCapabilities(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 6, len(c.callableCapabilities))
		assert.Equal(t, 1, len(c.nestedCapabilities))
	})

	t.Run("with pike contact capabilities - pike contact service is not registered -> should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()

		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))

		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithP2PCapabilities(),
			WithPikeContactCapabilities(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 7, len(c.callableCapabilities))
	})

	t.Run("with pike payment capabilities - pike payment service is not registered -> should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()

		sl := &PaymailServiceLocator{}
		sl.RegisterPaymailService(new(mockServiceProvider))

		c, err := NewConfig(
			sl,
			WithDomain("test.com"),
			WithP2PCapabilities(),
			WithPikePaymentCapabilities(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, 6, len(c.callableCapabilities))
	})
}
