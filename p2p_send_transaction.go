package paymail

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bsv-blockchain/go-paymail/beef"
)

var (
	// ErrSendP2PInvalidURL is returned when P2P URL is invalid
	ErrSendP2PInvalidURL = errors.New("invalid url")
	// ErrSendP2PMissingAlias is returned when alias is missing
	ErrSendP2PMissingAlias = errors.New("missing alias")
	// ErrSendP2PMissingDomain is returned when domain is missing
	ErrSendP2PMissingDomain = errors.New("missing domain")
	// ErrSendP2PTransactionNil is returned when transaction is nil
	ErrSendP2PTransactionNil = errors.New("transaction cannot be nil")
	// ErrSendP2PBeefOrHexRequired is returned when neither beef nor hex is provided
	ErrSendP2PBeefOrHexRequired = errors.New("beef or hex is required")
	// ErrSendP2PReferenceRequired is returned when reference is required but missing
	ErrSendP2PReferenceRequired = errors.New("reference is required")
	// ErrSendP2PAddressNotFound is returned when paymail address is not found
	ErrSendP2PAddressNotFound = errors.New("paymail address not found")
	// ErrSendP2PBadResponse is returned when receiving bad response from paymail provider
	ErrSendP2PBadResponse = errors.New("bad response from paymail provider")
	// ErrSendP2PMissingTxID is returned when returned txid is missing
	ErrSendP2PMissingTxID = errors.New("missing a returned txid")
)

/*
Example:
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

// P2PTransaction is the request body for the P2P transaction request
type P2PTransaction struct {
	Hex         string            `json:"hex"`         // The raw transaction, encoded as a hexadecimal string
	Beef        string            `json:"beef"`        // The transaction in hex BEEF format
	DecodedBeef *beef.DecodedBEEF `json:"decodedBeef"` // Decoded BEEF transaction
	MetaData    *P2PMetaData      `json:"metadata"`    // An object containing data associated with the transaction
	Reference   string            `json:"reference"`   // Reference for the payment (from previous P2P Destination request)
}

// P2PMetaData is an object containing data associated with the P2P transaction
type P2PMetaData struct {
	Note      string `json:"note,omitempty"`      // A human-readable bit of information about the payment
	PublicKey string `json:"pubkey,omitempty"`    // Public key to validate the signature (if signature is given)
	Sender    string `json:"sender,omitempty"`    // The paymail of the person that originated the transaction
	Signature string `json:"signature,omitempty"` // A signature of the tx id made by the sender
}

// P2PTransactionResponse is the response to the request
type P2PTransactionResponse struct {
	StandardResponse
	P2PTransactionPayload
}

// P2PTransactionPayload is payload from the request
type P2PTransactionPayload struct {
	Note string `json:"note"` // Some human-readable note
	TxID string `json:"txid"` // The txid of the broadcasted tx
}

// SendP2PTransaction will submit a transaction hex string (tx_hex) to a paymail provider
//
// Specs: https://docs.moneybutton.com/docs/paymail-06-p2p-transactions.html
func (c *Client) SendP2PTransaction(p2pURL, alias, domain string,
	transaction *P2PTransaction,
) (response *P2PTransactionResponse, err error) {
	// Require a valid url
	if len(p2pURL) == 0 || !strings.Contains(p2pURL, "https://") {
		err = fmt.Errorf("%s: %s: %w", "invalid url", p2pURL, ErrSendP2PInvalidURL)
		return response, err
	} else if len(alias) == 0 {
		err = ErrSendP2PMissingAlias
		return response, err
	} else if len(domain) == 0 {
		err = ErrSendP2PMissingDomain
		return response, err
	}

	// Basic requirements for request
	if transaction == nil {
		err = ErrSendP2PTransactionNil
		return response, err
	} else if len(transaction.Beef) == 0 && len(transaction.Hex) == 0 {
		err = ErrSendP2PBeefOrHexRequired
		return response, err
	} else if len(transaction.Reference) == 0 {
		err = ErrSendP2PReferenceRequired
		return response, err
	}

	// Set the base url and path, assuming the url is from the prior GetCapabilities() request
	// https://<host-discovery-target>/api/rawtx/{alias}@{domain.tld}
	// https://<host-discovery-target>/api/receive-transaction/{alias}@{domain.tld}
	reqURL := replaceAliasDomain(p2pURL, alias, domain)

	// Fire the POST request
	var resp StandardResponse
	if resp, err = c.postRequest(reqURL, transaction); err != nil {
		return response, err
	}

	// Start the response
	response = &P2PTransactionResponse{StandardResponse: resp}

	// Test the status code
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotModified {

		// Paymail address not found?
		if response.StatusCode == http.StatusNotFound {
			err = ErrSendP2PAddressNotFound
		} else {
			serverError := &ServerError{}
			if err = json.Unmarshal(resp.Body, serverError); err != nil {
				return response, err
			}
			if len(serverError.Message) == 0 {
				err = fmt.Errorf("code %d, body: %s: %w", response.StatusCode, string(resp.Body), ErrSendP2PBadResponse)
			} else {
				err = fmt.Errorf("code %d, message: %s: %w", response.StatusCode, serverError.Message, ErrSendP2PBadResponse)
			}
		}

		return response, err
	}

	// Decode the body of the response
	if err = json.Unmarshal(resp.Body, &response); err != nil {
		return response, err
	}

	// Check for a TX ID
	if len(response.TxID) == 0 {
		err = ErrSendP2PMissingTxID
		return response, err
	}

	return response, err
}
