package server

import (
	"github.com/gin-gonic/gin"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/bsv-blockchain/go-paymail/errors"
)

// GetPaymailAndCreateMetadata is a helper function to get the paymail from the request, check it in database and create the metadata based on that.
func (c *Configuration) GetPaymailAndCreateMetadata(context *gin.Context, satoshis uint64) (alias, domain string, md *RequestMetadata, ok bool) {
	incomingPaymail := context.Param(PaymailAddressParamName)

	// Parse, sanitize and basic validation
	alias, domain, paymailAddress := paymail.SanitizePaymail(incomingPaymail)
	if len(paymailAddress) == 0 {
		errors.ErrorResponse(context, errors.ErrInvalidPaymail, c.Logger)
		return alias, domain, md, ok
	}
	if !c.IsAllowedDomain(domain) {
		errors.ErrorResponse(context, errors.ErrDomainUnknown, c.Logger)
		return alias, domain, md, ok
	}

	// Start the PaymentRequest
	paymentRequest := &paymail.PaymentRequest{
		Satoshis: satoshis,
	}

	// Did we get some satoshis?
	if paymentRequest.Satoshis == 0 {
		errors.ErrorResponse(context, errors.ErrMissingFieldSatoshis, c.Logger)
		return alias, domain, md, ok
	}

	// Create the metadata struct
	md = CreateMetadata(context.Request, alias, domain, "")
	md.PaymentDestination = paymentRequest

	// Get from the data layer
	foundPaymail, err := c.actions.GetPaymailByAlias(context.Request.Context(), alias, domain, md)
	if err != nil {
		errors.ErrorResponse(context, err, c.Logger)
		return alias, domain, md, ok
	}
	if foundPaymail == nil {
		errors.ErrorResponse(context, errors.ErrCouldNotFindPaymail, c.Logger)
		return alias, domain, md, ok
	}

	ok = true
	return alias, domain, md, ok
}
