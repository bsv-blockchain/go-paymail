package server

import (
	"encoding/json"
	"net/http"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/bsv-blockchain/go-paymail/errors"
)

func parseP2pReceiveTxRequest(c *Configuration, req *http.Request, incomingPaymail string, format p2pPayloadFormat) (*p2pReceiveTxReqPayload, error) {
	alias, domain, paymailAddress := paymail.SanitizePaymail(incomingPaymail)
	if len(paymailAddress) == 0 {
		return nil, errors.ErrInvalidPaymail
	} else if !c.IsAllowedDomain(domain) {
		return nil, errors.ErrDomainUnknown
	}

	requestData := p2pReceiveTxReqPayload{
		incomingPaymailAlias:  alias,
		incomingPaymailDomain: domain,
	}

	var p2pTransaction paymail.P2PTransaction
	err := json.NewDecoder(req.Body).Decode(&p2pTransaction)
	if err != nil {
		return nil, errors.ErrCannotBindRequest
	}
	if len(p2pTransaction.Reference) == 0 {
		return nil, errors.ErrMissingFieldReference
	}
	switch format {
	case basicP2pPayload:
		if len(p2pTransaction.Hex) == 0 {
			return nil, errors.ErrMissingFieldHex
		}
	case beefP2pPayload:
		if len(p2pTransaction.Beef) == 0 {
			return nil, errors.ErrMissingFieldBEEF
		}
	}
	vErr := validateMetadata(c, p2pTransaction.MetaData)

	if vErr != nil {
		return nil, vErr
	}

	requestData.P2PTransaction = &p2pTransaction
	return &requestData, nil
}

func validateMetadata(c *Configuration, metadata *paymail.P2PMetaData) error {
	// Check signature if: 1) sender validation enabled or 2) a signature was given (optional)
	if c.SenderValidationEnabled || len(metadata.Signature) > 0 {

		// Check required fields for signature validation
		if len(metadata.Signature) == 0 {
			return errors.ErrMissingFieldSignature
		}

		if len(metadata.PublicKey) == 0 {
			return errors.ErrMissingFieldPubKey
		}
	}

	return nil
}
