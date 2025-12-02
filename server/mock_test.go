package server

import (
	"context"
	"errors"

	"github.com/bsv-blockchain/go-paymail"
	"github.com/bsv-blockchain/go-paymail/spv"
)

var errMockNotImplemented = errors.New("mock not implemented")

// Mock implementation of a service provider
type mockServiceProvider struct{}

// GetPaymailByAlias is a demo implementation of this interface
func (m *mockServiceProvider) GetPaymailByAlias(_ context.Context, _, _ string,
	_ *RequestMetadata,
) (*paymail.AddressInformation, error) {
	// Get the data from the demo database
	return nil, errMockNotImplemented
}

// CreateAddressResolutionResponse is a demo implementation of this interface
func (m *mockServiceProvider) CreateAddressResolutionResponse(_ context.Context, _, _ string,
	_ bool, _ *RequestMetadata,
) (*paymail.ResolutionPayload, error) {
	// Generate a new destination / output for the basic address resolution
	return nil, errMockNotImplemented
}

// CreateP2PDestinationResponse is a demo implementation of this interface
func (m *mockServiceProvider) CreateP2PDestinationResponse(_ context.Context, _, _ string,
	_ uint64, _ *RequestMetadata,
) (*paymail.PaymentDestinationPayload, error) {
	// Generate a new destination for the p2p request
	return nil, errMockNotImplemented
}

// RecordTransaction is a demo implementation of this interface
func (m *mockServiceProvider) RecordTransaction(_ context.Context,
	_ *paymail.P2PTransaction, _ *RequestMetadata,
) (*paymail.P2PTransactionPayload, error) {
	// Record the tx into your datastore layer
	return nil, errMockNotImplemented
}

// VerifyMerkleRoots is a mock implementation of this interface
func (m *mockServiceProvider) VerifyMerkleRoots(_ context.Context, _ []*spv.MerkleRootConfirmationRequestItem) error {
	// Verify the merkle roots
	return nil
}

func (m *mockServiceProvider) AddContact(ctx context.Context, requesterPaymail string, contact *paymail.PikeContactRequestPayload) error {
	return nil
}

func (m *mockServiceProvider) CreatePikeOutputResponse(ctx context.Context, alias, domain, senderPubKey string, satoshis uint64, metaData *RequestMetadata) (*paymail.PikePaymentOutputsResponse, error) {
	return nil, errMockNotImplemented
}
