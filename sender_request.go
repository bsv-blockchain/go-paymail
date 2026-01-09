package paymail

import (
	"errors"
	"fmt"

	bsm "github.com/bsv-blockchain/go-sdk/compat/bsm"
	primitives "github.com/bsv-blockchain/go-sdk/primitives/ec"
)

var (
	// ErrSenderRequestMissingKeyAddress is returned when key address is missing
	ErrSenderRequestMissingKeyAddress = errors.New("missing key address")
	// ErrSenderRequestMissingSignature is returned when signature is missing
	ErrSenderRequestMissingSignature = errors.New("missing a signature to verify")
	// ErrSenderRequestMissingPrivateKey is returned when private key is missing
	ErrSenderRequestMissingPrivateKey = errors.New("missing private key")
	// ErrSenderRequestMissingDt is returned when dt is missing
	ErrSenderRequestMissingDt = errors.New("missing dt")
	// ErrSenderRequestMissingSenderHandle is returned when senderHandle is missing
	ErrSenderRequestMissingSenderHandle = errors.New("missing senderHandle")
)

/*
Example:
{
    "senderName": "FirstName LastName",
    "senderHandle": "<alias>@<domain.tld>",
    "dt": "2013-10-21T13:28:06.419Z",
    "amount": 550,
    "purpose": "message to receiver",
    "signature": "<compact Bitcoin message signature>"
}
*/

// SenderRequest is the request body for the basic address resolution
//
// This is required to make a basic resolution request, and Dt and SenderHandle are required
type SenderRequest struct {
	Amount       uint64 `json:"amount,omitempty"`     // The amount, in Satoshis, that the sender intends to transfer to the receiver
	Dt           string `json:"dt"`                   // (required) ISO-8601 formatted timestamp; see notes
	Purpose      string `json:"purpose,omitempty"`    // Human-readable description of the purpose of the payment
	SenderHandle string `json:"senderHandle"`         // (required) Sender's paymail handle
	SenderName   string `json:"senderName,omitempty"` // Human-readable sender display name
	Signature    string `json:"signature,omitempty"`  // Compact Bitcoin message signature; http://bsvalias.org/04-01-basic-address-resolution.html#signature-field
}

// Verify will verify the given components in the ResolveAddress() request
//
// Source: https://github.com/moneybutton/paymail-client/blob/master/src/VerifiableMessage.js
// Specs: http://bsvalias.org/04-01-basic-address-resolution.html#signature-field
func (s *SenderRequest) Verify(keyAddress, signature string) error {
	// Basic checks before trying the signature verification
	if len(keyAddress) == 0 {
		return ErrSenderRequestMissingKeyAddress
	} else if len(signature) == 0 {
		return ErrSenderRequestMissingSignature
	}

	decodedSig, err := DecodeSignature(signature)
	if err != nil {
		return err
	}

	// Concatenate & verify the message
	return bsm.VerifyMessage(keyAddress, decodedSig, prepareMessage(s))
}

// Sign will sign the given components in the ResolveAddress() request
//
// Source: https://github.com/moneybutton/paymail-client/blob/master/src/VerifiableMessage.js
// Specs: http://bsvalias.org/04-01-basic-address-resolution.html#signature-field
// Additional Specs: http://bsvalias.org/04-02-sender-validation.html
func (s *SenderRequest) Sign(privateKey string) ([]byte, error) {
	// Basic checks before trying to sign the request
	if len(privateKey) == 0 {
		return nil, ErrSenderRequestMissingPrivateKey
	} else if len(s.Dt) == 0 {
		return nil, ErrSenderRequestMissingDt
	} else if len(s.SenderHandle) == 0 {
		return nil, ErrSenderRequestMissingSenderHandle
	}

	privKey, err := primitives.PrivateKeyFromHex(privateKey)
	if err != nil {
		return nil, err
	}

	// Concatenate & sign message
	return bsm.SignMessage(
		privKey,
		prepareMessage(s),
	)
}

func prepareMessage(senderRequest *SenderRequest) []byte {
	return []byte(fmt.Sprintf("%s%d%s%s", senderRequest.SenderHandle, senderRequest.Amount, senderRequest.Dt, senderRequest.Purpose))
}
