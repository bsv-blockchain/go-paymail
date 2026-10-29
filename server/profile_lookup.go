package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/bsv-blockchain/go-paymail/errors"
)

// SetProfileLookupCapabilities adds the BRC-235 public profile lookup and reverse lookup
// capabilities. They need no other capability, so a server configured with
// WithoutGenericCapabilities() and WithProfileLookupCapabilities() advertises these two alone.
func (c *Configuration) SetProfileLookupCapabilities() {
	_addCapabilities(
		c.callableCapabilities,
		CallableCapabilitiesMap{
			paymail.BRFCProfileLookup: CallableCapability{
				Path:    "/handle/" + ProfileQueryTemplate,
				Method:  http.MethodGet,
				Handler: c.profileLookup,
			},
			paymail.BRFCProfileReverseLookup: CallableCapability{
				Path:    "/identityKey/" + PubKeyTemplate,
				Method:  http.MethodGet,
				Handler: c.profileReverseLookup,
			},
		},
	)
}

// profileDomain is the paymail domain a profile lookup request is for: the request host
// when it is an allowed domain, otherwise the single configured domain. A lookup carries
// no paymail in its path, and a host reached through SRV is not the paymail domain.
func (c *Configuration) profileDomain(context *gin.Context) (string, bool) {
	host, err := paymail.SanitizeDomain(context.Request.Host)
	if err == nil && host != "" && (c.PaymailDomainsValidationDisabled || c.IsAllowedDomain(host)) {
		return host, true
	}
	if !c.PaymailDomainsValidationDisabled && len(c.PaymailDomains) == 1 {
		return strings.ToLower(c.PaymailDomains[0].Name), true
	}
	return "", false
}

// profileLookup answers a handle query with profile certificates
//
// Specs: https://github.com/bsv-blockchain/BRCs/blob/master/payments/0235.md
func (c *Configuration) profileLookup(context *gin.Context) {
	domain, ok := c.profileDomain(context)
	if !ok {
		errors.ErrorResponse(context, errors.ErrDomainUnknown, c.Logger)
		return
	}
	certificates := make([]json.RawMessage, 0, paymail.ProfileLookupMaxResults)
	query, ok := paymail.NormalizeProfileLookupQuery(context.Param(ProfileQueryParamName), domain)
	if !ok {
		context.JSON(http.StatusOK, certificates)
		return
	}
	results, err := c.profileLookupActions.LookupProfiles(context.Request.Context(), domain, query)
	if err != nil {
		errors.ErrorResponse(context, err, c.Logger)
		return
	}
	for _, raw := range results {
		if len(certificates) >= paymail.ProfileLookupMaxResults {
			break
		}
		if _, checkErr := paymail.CheckProfileCertificate(raw, domain); checkErr == nil {
			certificates = append(certificates, raw)
		}
	}
	context.JSON(http.StatusOK, certificates)
}

// profileReverseLookup answers the profile certificate an identity key holds here
//
// Specs: https://github.com/bsv-blockchain/BRCs/blob/master/payments/0235.md
func (c *Configuration) profileReverseLookup(context *gin.Context) {
	domain, ok := c.profileDomain(context)
	if !ok {
		errors.ErrorResponse(context, errors.ErrDomainUnknown, c.Logger)
		return
	}
	identityKey := strings.ToLower(context.Param(PubKeyParamName))
	if !paymail.IsProfileIdentityKey(identityKey) {
		errors.ErrorResponse(context, errors.ErrInvalidPubKey, c.Logger)
		return
	}
	raw, err := c.profileLookupActions.ReverseLookupProfile(context.Request.Context(), domain, identityKey)
	if err != nil {
		errors.ErrorResponse(context, err, c.Logger)
		return
	}
	if raw == nil {
		errors.ErrorResponse(context, errors.ErrProfileNotFound, c.Logger)
		return
	}
	cert, err := paymail.CheckProfileCertificate(raw, domain)
	if err != nil || strings.ToLower(cert.Subject) != identityKey {
		c.Logger.Error().Msg("profile reverse lookup provider returned an unservable certificate")
		errors.ErrorResponse(context, errors.ErrProfileUnservable, c.Logger)
		return
	}
	context.Data(http.StatusOK, "application/json; charset=utf-8", raw)
}
