package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	bsm "github.com/bitcoin-sv/go-sdk/compat/bsm"
	ec "github.com/bitcoin-sv/go-sdk/primitives/ec"
	"github.com/bitcoin-sv/go-sdk/script"
	"github.com/bitcoin-sv/go-sdk/transaction/template/p2pkh"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/bsv-blockchain/go-paymail/logging"
)

var (
	// ErrDemoUnableDecodePrivateKey is returned when unable to decode private key
	ErrDemoUnableDecodePrivateKey = errors.New("unable to decode private key")
	// ErrDemoInvalidSignature is returned when signature is invalid
	ErrDemoInvalidSignature = errors.New("invalid signature")
	// ErrDemoFailedCreatePaymail is returned when failed to create paymail address in database
	ErrDemoFailedCreatePaymail = errors.New("failed to create paymail address in demo database")
	// ErrDemoAddressNotFound is returned when paymail address is not found
	ErrDemoAddressNotFound = errors.New("address not found")
)

// paymailAddressTable is the demo data for the example server (table: paymail_address)
var demoPaymailAddressTable []*paymail.AddressInformation

// Create the list of demo aliases to create on load
var demoAliases = []struct {
	alias  string
	domain string
	avatar string
	id     string
	name   string
}{
	{"mrz", "test.com", "https://github.com/mrz1836.png", "1", "MrZ"},
	{"mrz", "another.com", "https://github.com/mrz1836.png", "4", "MrZ"},
	{"satchmo", "test.com", "https://github.com/rohenaz.png", "2", "Satchmo"},
	{"siggi", "test.com", "https://github.com/icellan.png", "3", "Siggi"},
}

// InitDemoDatabase creates demo data for the database based on the given aliases
func InitDemoDatabase() error {
	// Generate paymail address records
	for _, demo := range demoAliases {
		if err := generateDemoPaymail(
			demo.alias,
			demo.domain,
			demo.avatar,
			demo.name,
			demo.id,
		); err != nil {
			return fmt.Errorf("alias: %s id: %s: %w", demo.alias, demo.id, ErrDemoFailedCreatePaymail)
		}
	}

	return nil
}

// generateDemoPaymail will make a new row in the demo database
// This function creates a private key and pubkey
func generateDemoPaymail(alias, domain, avatar, name, id string) (err error) {
	// Start a row
	row := &paymail.AddressInformation{
		Alias:  alias,
		Avatar: avatar,
		Domain: domain,
		ID:     id,
		Name:   name,
	}

	// Generate new private key
	key, err := ec.NewPrivateKey()
	if err != nil {
		return err
	}

	row.PrivateKey = hex.EncodeToString((key.Serialize()))

	addr, err := script.NewAddressFromPublicKey(key.PubKey(), true)
	if err != nil {
		return err
	}

	row.LastAddress = addr.AddressString

	// Add to the table
	demoPaymailAddressTable = append(demoPaymailAddressTable, row)

	return err
}

// DemoGetPaymailByAlias will find a paymail address given an alias
func DemoGetPaymailByAlias(alias, domain string) (*paymail.AddressInformation, error) {
	for i, row := range demoPaymailAddressTable {
		if strings.EqualFold(alias, row.Alias) && strings.EqualFold(domain, row.Domain) {
			return demoPaymailAddressTable[i], nil
		}
	}
	return nil, ErrDemoAddressNotFound
}

// DemoCreateAddressResolutionResponse will create a new destination for the address resolution
func DemoCreateAddressResolutionResponse(_ context.Context, alias, domain string,
	senderValidation bool,
) (*paymail.ResolutionPayload, error) {
	// Get the paymail record
	p, err := DemoGetPaymailByAlias(alias, domain)
	if err != nil {
		return nil, err
	}

	// Start the response
	response := &paymail.ResolutionPayload{}

	// Generate the script
	sc, _ := script.NewAddressFromString(p.LastAddress)
	ls, _ := p2pkh.Lock(sc)

	response.Output = ls.String()

	privateKeyFromHex, err := ec.PrivateKeyFromHex(p.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to decode private key: %w", err)
	}

	// Create a signature of output if senderValidation is enabled
	if senderValidation {
		sigBytes, err := bsm.SignMessage(privateKeyFromHex, ls.Bytes())
		if err != nil {
			return nil, fmt.Errorf("invalid signature: %w", err)
		}
		response.Signature = paymail.EncodeSignature(sigBytes)
	}

	return response, nil
}

// DemoCreateP2PDestinationResponse will create a basic resolution response for the demo
func DemoCreateP2PDestinationResponse(_ context.Context, alias, domain string,
	satoshis uint64,
) (*paymail.PaymentDestinationPayload, error) {
	// Get the paymail record
	p, err := DemoGetPaymailByAlias(alias, domain)
	if err != nil {
		return nil, err
	}

	// Start the output
	output := &paymail.PaymentOutput{
		Satoshis: satoshis,
	}

	sc, _ := script.NewAddressFromString(p.LastAddress)
	ls, _ := p2pkh.Lock(sc)
	output.Script = ls.String()

	// Create the response
	return &paymail.PaymentDestinationPayload{
		Outputs:   []*paymail.PaymentOutput{output},
		Reference: "1234567890", // todo: this should be unique per request
	}, nil
}

// DemoRecordTransaction will record the tx in the datalayer
func DemoRecordTransaction(_ context.Context,
	p2pTx *paymail.P2PTransaction,
) (*paymail.P2PTransactionPayload, error) {
	logger := logging.GetDefaultLogger()

	// Record the transaction
	logger.Info().Msgf("recording tx... reference: %s\n", p2pTx.Reference)

	// Broadcast etc...

	// Convert the hex to TxID
	/*
		tx, err := bt.NewTxFromString(p2pTx.Hex)
		if err != nil {
			return nil, err
		}
	*/

	// Creating a FAKE tx id for this demo
	hash := sha256.Sum256([]byte(p2pTx.Hex))
	fakeTxID := hex.EncodeToString(hash[:])

	// Demo response
	return &paymail.P2PTransactionPayload{
		Note: p2pTx.MetaData.Note,
		TxID: fakeTxID,
	}, nil
}
