package server

import (
	"context"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/bsv-blockchain/go-paymail/spv"
)

type PaymailServiceLocator struct {
	paymailService     PaymailServiceProvider
	pikeContactService PikeContactServiceProvider
	pikePaymentService PikePaymentServiceProvider
}

func (l *PaymailServiceLocator) RegisterPaymailService(s PaymailServiceProvider) {
	l.paymailService = s
}

func (l *PaymailServiceLocator) GetPaymailService() PaymailServiceProvider {
	if l.paymailService == nil {
		panic("PaymailServiceProvider was not registered")
	}

	return l.paymailService
}

func (l *PaymailServiceLocator) RegisterPikeContactService(s PikeContactServiceProvider) {
	l.pikeContactService = s
}

func (l *PaymailServiceLocator) GetPikeContactService() PikeContactServiceProvider {
	if l.pikeContactService == nil {
		panic("PikeContactServiceProvider was not registered")
	}

	return l.pikeContactService
}

func (l *PaymailServiceLocator) RegisterPikePaymentService(s PikePaymentServiceProvider) {
	l.pikePaymentService = s
}

func (l *PaymailServiceLocator) GetPikePaymentService() PikePaymentServiceProvider {
	if l.pikePaymentService == nil {
		panic("PikePaymentServiceProvider was not registered")
	}

	return l.pikePaymentService
}

// PaymailServiceProvider the paymail server interface that needs to be implemented
type PaymailServiceProvider interface {
	CreateAddressResolutionResponse(
		ctx context.Context,
		alias, domain string,
		senderValidation bool,
		metaData *RequestMetadata,
	) (*paymail.ResolutionPayload, error)

	CreateP2PDestinationResponse(
		ctx context.Context,
		alias, domain string,
		satoshis uint64,
		metaData *RequestMetadata,
	) (*paymail.PaymentDestinationPayload, error)

	GetPaymailByAlias(
		ctx context.Context,
		alias, domain string,
		metaData *RequestMetadata,
	) (*paymail.AddressInformation, error)

	RecordTransaction(
		ctx context.Context,
		p2pTx *paymail.P2PTransaction,
		metaData *RequestMetadata,
	) (*paymail.P2PTransactionPayload, error)

	VerifyMerkleRoots(
		ctx context.Context,
		merkleProofs []*spv.MerkleRootConfirmationRequestItem,
	) error
}

type PikeContactServiceProvider interface {
	AddContact(
		ctx context.Context,
		requesterPaymail string,
		contact *paymail.PikeContactRequestPayload,
	) error
}

type PikePaymentServiceProvider interface {
	CreatePikeOutputResponse(
		ctx context.Context,
		alias, domain, senderPubKey string,
		satoshis uint64,
		metaData *RequestMetadata,
	) (*paymail.PikePaymentOutputsResponse, error)
}
