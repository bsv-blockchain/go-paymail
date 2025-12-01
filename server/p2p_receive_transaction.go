package server

import (
	"net/http"

	"github.com/bsv-blockchain/go-paymail/errors"
	"github.com/gin-gonic/gin"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/bsv-blockchain/go-paymail/spv"
)

type p2pPayloadFormat uint

const (
	basicP2pPayload p2pPayloadFormat = iota
	beefP2pPayload
)

/*
Incoming Data Object Example:
{
  "hex": "01000000012adda020db81f2155ebba69e7.........154888ac00000000",
  "metadata": {
	"sender": "someone@example.tld",
	"pubkey": "<sender-pubkey>",
	"signature": "signature(txid)",
	"note": "Human readable information related to the tx."
  },
  "reference": "someRefId"
}
*/

// p2pReceiveTx will receive a P2P transaction (from previous request: P2P Payment Destination)
//
// Specs: https://docs.moneybutton.com/docs/paymail-06-p2p-transactions.html
func (c *Configuration) p2pReceiveTx(context *gin.Context) {
	p2pFormat := basicP2pPayload

	incomingPaymail := context.Param(PaymailAddressParamName)

	requestPayload, _, md, err := processP2pReceiveTxRequest(c, context.Request, incomingPaymail, p2pFormat)
	if err != nil {
		errors.ErrorResponse(context, err, c.Logger)
		return
	}

	if len(requestPayload.Hex) == 0 {
		panic("empty hex after parsing!")
	}

	var response *paymail.P2PTransactionPayload
	if response, err = c.actions.RecordTransaction(
		context.Request.Context(), requestPayload.P2PTransaction, md,
	); err != nil {
		errors.ErrorResponse(context, err, c.Logger)
		return
	}

	context.JSON(http.StatusOK, response)
}

/*
Incoming Data Object Example:
{
  "beef": "01000000012adda020db81f2155ebba69e7.........154888ac00000000",
  "metadata": {
	"sender": "someone@example.tld",
	"pubkey": "<sender-pubkey>",
	"signature": "signature(txid)",
	"note": "Human readable information related to the tx."
  },
  "reference": "someRefId"
}
*/
// p2pReceiveBeefTx will receive a P2P transaction in BEEF format
func (c *Configuration) p2pReceiveBeefTx(context *gin.Context) {
	p2pFormat := beefP2pPayload
	incomingPaymail := context.Param(PaymailAddressParamName)

	requestPayload, dBeef, md, err := processP2pReceiveTxRequest(c, context.Request, incomingPaymail, p2pFormat)
	if err != nil {
		errors.ErrorResponse(context, err, c.Logger)
		return
	}

	if len(requestPayload.Hex) == 0 {
		panic("empty hex after parsing!")
	}

	if dBeef == nil {
		panic("empty beef after parsing!")
	}

	err = spv.ExecuteSimplifiedPaymentVerification(context.Request.Context(), dBeef, c.actions)
	if err != nil {
		errors.ErrorResponse(context, errors.ErrSPVFailed, c.Logger)
		return
	}

	var response *paymail.P2PTransactionPayload
	if response, err = c.actions.RecordTransaction(
		context.Request.Context(), requestPayload.P2PTransaction, md,
	); err != nil {
		errors.ErrorResponse(context, err, c.Logger)
		return
	}

	context.JSON(http.StatusOK, response)
}
