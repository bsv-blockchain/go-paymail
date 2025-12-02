package server

import (
	"net"
	"net/http"
	"time"

	ec "github.com/bitcoin-sv/go-sdk/primitives/ec"
	script "github.com/bitcoin-sv/go-sdk/script"
	"github.com/gin-gonic/gin"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/bsv-blockchain/go-paymail/errors"
)

// TODO: bitcoin.PubKeyFromString -> PubKeyFromSignature?
// TODO: bitcoin.GetAddressFromPubKey -> NewAddressFromPublicKeyString?

/*
Incoming Data Object Example:
{
    "senderName": "UserName",
    "senderHandle": "alias@domain.com",
    "dt": "2020-04-09T16:08:06.419Z",
    "amount": 551,
    "purpose": "message to receiver",
	"signature": "SIGNATURE-IF-REQUIRED-IN-CONFIG"
}
*/

// resolveAddress will return the payment destination (bitcoin address) for the corresponding paymail address
//
// Specs: http://bsvalias.org/04-01-basic-address-resolution.html
func (c *Configuration) resolveAddress(context *gin.Context) {
	incomingPaymail := context.Param(PaymailAddressParamName)

	// Parse, sanitize and basic validation
	alias, domain, paymailAddress := paymail.SanitizePaymail(incomingPaymail)
	if len(paymailAddress) == 0 {
		context.JSON(http.StatusBadRequest, "invalid paymail: "+incomingPaymail)
		errors.ErrorResponse(context, errors.ErrInvalidPaymail, c.Logger)
		return
	} else if !c.IsAllowedDomain(domain) {
		errors.ErrorResponse(context, errors.ErrDomainUnknown, c.Logger)
		return
	}

	var senderRequest paymail.SenderRequest
	err := context.Bind(&senderRequest)
	if err != nil {
		errors.ErrorResponse(context, errors.ErrCannotBindRequest, c.Logger)
		return
	}

	// Check for required fields
	if len(senderRequest.SenderHandle) == 0 {
		errors.ErrorResponse(context, errors.ErrSenderHandleEmpty, c.Logger)
		return
	} else if len(senderRequest.Dt) == 0 {
		errors.ErrorResponse(context, errors.ErrDtEmpty, c.Logger)
		return
	}

	// Validate the timestamp
	if err = paymail.ValidateTimestamp(senderRequest.Dt); err != nil {
		errors.ErrorResponse(context, errors.ErrInvalidTimestamp, c.Logger)
		return
	}

	// Basic validation on sender handle
	if err = paymail.ValidatePaymail(senderRequest.SenderHandle); err != nil {
		errors.ErrorResponse(context, errors.ErrInvalidSenderHandle, c.Logger)
		return
	}

	// Only validate signatures if sender validation is enabled (skip if disabled)
	if c.SenderValidationEnabled {
		if len(senderRequest.Signature) > 0 {

			// Get the pubKey from the corresponding sender paymail address
			var senderPubKey *ec.PublicKey
			senderPubKey, err = getSenderPubKey(senderRequest.SenderHandle)
			if err != nil {
				errors.ErrorResponse(context, err, c.Logger)
				return
			}

			// Derive address from pubKey
			var rawAddress *script.Address
			if rawAddress, err = script.NewAddressFromPublicKey(senderPubKey, true); err != nil {
				errors.ErrorResponse(context, errors.ErrInvalidSenderHandle, c.Logger)
				return
			}

			// Verify the signature
			if err = senderRequest.Verify(rawAddress.AddressString, senderRequest.Signature); err != nil {
				errors.ErrorResponse(context, errors.ErrInvalidSignature, c.Logger)
				return
			}
		} else {
			errors.ErrorResponse(context, errors.ErrMissingFieldSignature, c.Logger)
			return
		}
	}

	// Create the metadata struct
	md := CreateMetadata(context.Request, alias, domain, "")
	md.ResolveAddress = &senderRequest

	// Get from the data layer
	foundPaymail, err := c.actions.GetPaymailByAlias(context.Request.Context(), alias, domain, md)
	if err != nil {
		errors.ErrorResponse(context, err, c.Logger)
		return
	} else if foundPaymail == nil {
		errors.ErrorResponse(context, errors.ErrCouldNotFindPaymail, c.Logger)
		return
	}

	// Get the resolution information
	var response *paymail.ResolutionPayload
	if response, err = c.actions.CreateAddressResolutionResponse(
		context.Request.Context(), alias, domain, c.SenderValidationEnabled, md,
	); err != nil {
		errors.ErrorResponse(context, err, c.Logger)
		return
	}

	// Set the response
	context.JSON(http.StatusOK, response)
}

// getSenderPubKey will fetch the pubKey from a PKI request for the sender handle
func getSenderPubKey(senderPaymailAddress string) (*ec.PublicKey, error) {
	// Sanitize and break apart
	alias, domain, _ := paymail.SanitizePaymail(senderPaymailAddress)

	// Load the client
	client, err := paymail.NewClient(paymail.WithHTTPTimeout(15 * time.Second))
	if err != nil {
		return nil, err
	}

	// Get the SRV record
	var srv *net.SRV
	if srv, err = client.GetSRVRecord(
		paymail.DefaultServiceName, paymail.DefaultProtocol, domain,
	); err != nil {
		return nil, err
	}

	// Get the capabilities
	// This is required first to get the corresponding PKI endpoint url
	var capabilities *paymail.CapabilitiesResponse
	if capabilities, err = client.GetCapabilities(
		srv.Target, paymail.DefaultPort,
	); err != nil {
		return nil, err
	}

	// Extract the PKI URL from the capabilities response
	pkiURL := capabilities.GetString(paymail.BRFCPki, paymail.BRFCPkiAlternate)

	// Get the actual PKI
	var pki *paymail.PKIResponse
	if pki, err = client.GetPKI(
		pkiURL, alias, domain,
	); err != nil {
		return nil, err
	}

	// Convert the string pubKey to a ec.PubKey
	return ec.PublicKeyFromString(pki.PubKey)
}
