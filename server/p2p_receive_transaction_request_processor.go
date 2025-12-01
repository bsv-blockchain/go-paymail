package server

import (
	"context"
	"net/http"

	bsm "github.com/bitcoin-sv/go-sdk/compat/bsm"
	script "github.com/bitcoin-sv/go-sdk/script"
	sdk "github.com/bitcoin-sv/go-sdk/transaction"
	"github.com/rs/zerolog"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/bsv-blockchain/go-paymail/beef"
	"github.com/bsv-blockchain/go-paymail/errors"
)

type p2pReceiveTxReqPayload struct {
	*paymail.P2PTransaction
	incomingPaymailAlias, incomingPaymailDomain string
}

func processP2pReceiveTxRequest(c *Configuration, req *http.Request, incomingPaymail string, format p2pPayloadFormat) (
	*p2pReceiveTxReqPayload, *beef.DecodedBEEF, *RequestMetadata, error,
) {
	payload, err := parseP2pReceiveTxRequest(c, req, incomingPaymail, format)
	if err != nil {
		return returnError(err)
	}

	md := CreateMetadata(req, payload.incomingPaymailAlias, payload.incomingPaymailDomain, "")
	err = verifyIncomingPaymail(req.Context(), c, md, payload.incomingPaymailAlias, payload.incomingPaymailDomain)
	if err != nil {
		return returnError(err)
	}

	tx, beefData, err := getProcessedTxData(payload, format, c.Logger)
	if err != nil {
		return returnError(err)
	}

	if c.SenderValidationEnabled || len(payload.MetaData.Signature) > 0 {
		err = verifySignature(payload.MetaData, tx.TxID().String())
		if err != nil {
			return returnError(err)
		}
	}

	if format == beefP2pPayload {
		payload.Hex = tx.String()
		payload.DecodedBeef = beefData
	}

	return payload, beefData, md, nil
}

func getProcessedTxData(payload *p2pReceiveTxReqPayload, format p2pPayloadFormat, log *zerolog.Logger) (*sdk.Transaction, *beef.DecodedBEEF, error) {
	var processedTx *sdk.Transaction
	var beefData *beef.DecodedBEEF
	var err error

	switch format {
	case basicP2pPayload:
		processedTx, err = sdk.NewTransactionFromHex(payload.Hex)
		if err != nil {
			log.Error().Msgf("error while parsing hex: %s", err.Error())
			return nil, nil, errors.ErrProcessingHex
		}

	case beefP2pPayload:
		beefData, err = beef.DecodeBEEF(payload.Beef)
		if err != nil {
			log.Error().Msgf("error while parsing beef: %s", err.Error())
			return nil, nil, errors.ErrProcessingBEEF
		}

		processedTx = beefData.GetLatestTx()

	default:
		panic("Unexpected transaction format!")
	}

	return processedTx, beefData, nil
}

func verifyIncomingPaymail(ctx context.Context, c *Configuration, md *RequestMetadata, alias, domain string) error {
	var foundPaymail *paymail.AddressInformation
	var err error

	foundPaymail, err = c.actions.GetPaymailByAlias(ctx, alias, domain, md)
	if err != nil {
		return err
	} else if foundPaymail == nil {
		return errors.ErrCouldNotFindPaymail
	}

	return nil
}

func verifySignature(metadata *paymail.P2PMetaData, txID string) error {
	// Get the address from pubKey
	var rawAddress *script.Address
	var err error

	if rawAddress, err = script.NewAddressFromPublicKeyString(metadata.PublicKey, true); err != nil {
		return errors.ErrInvalidPubKey
	}

	// Validate the signature of the tx id
	if err = bsm.VerifyMessage(rawAddress.AddressString, []byte(metadata.Signature), []byte(txID)); err != nil {
		return errors.ErrInvalidSignature
	}

	return nil
}

func returnError(err error) (
	*p2pReceiveTxReqPayload, *beef.DecodedBEEF, *RequestMetadata, error,
) {
	return nil, nil, nil, err
}
