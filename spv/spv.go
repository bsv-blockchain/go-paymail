package spv

import (
	"context"

	sdk "github.com/bitcoin-sv/go-sdk/transaction"

	"github.com/bsv-blockchain/go-paymail/beef"
	"github.com/bsv-blockchain/go-paymail/errors"
)

type MerkleRootVerifier interface {
	VerifyMerkleRoots(
		ctx context.Context,
		merkleRoots []*MerkleRootConfirmationRequestItem,
	) error
}

// MerkleRootConfirmationRequestItem is a request type for verification
// of Merkle Roots inclusion in the longest chain.
type MerkleRootConfirmationRequestItem struct {
	MerkleRoot  string `json:"merkleRoot"`
	BlockHeight uint64 `json:"blockHeight"`
}

// ExecuteSimplifiedPaymentVerification executes the SPV for decoded BEEF tx
func ExecuteSimplifiedPaymentVerification(ctx context.Context, dBeef *beef.DecodedBEEF, provider MerkleRootVerifier) error {
	for _, txDt := range dBeef.Transactions {
		tx := txDt.Transaction

		if len(tx.Outputs) == 0 {
			return errors.ErrNoOutputs
		}

		if len(tx.Inputs) == 0 {
			return errors.ErrNoInputs
		}

		if err := validateLockTime(tx); err != nil {
			return err
		}

		if txDt.Unmined() {
			if err := validateSatoshisSum(tx, dBeef.Transactions); err != nil {
				return err
			}

			if err := validateScripts(tx, dBeef.Transactions); err != nil {
				return err
			}
		}
	}

	if err := verifyMerkleRoots(ctx, dBeef, provider); err != nil {
		return err
	}

	return nil
}

func validateLockTime(tx *sdk.Transaction) error {
	if tx.LockTime == 0 {
		return nil
	}
	for _, input := range tx.Inputs {
		if input.SequenceNumber != 0xffffffff {
			return errors.ErrLockTimeAndSequence
		}
	}
	return nil
}

func validateSatoshisSum(tx *sdk.Transaction, inputTxs []*beef.TxData) error {
	inputSum, outputSum := uint64(0), uint64(0)

	for _, input := range tx.Inputs {
		inputParentTx := findParentForInput(input, inputTxs)

		if inputParentTx == nil {
			return errors.ErrInvalidParentTransactions
		}

		inputSum += inputParentTx.Transaction.Outputs[input.SourceTxOutIndex].Satoshis
	}
	for _, output := range tx.Outputs {
		outputSum += output.Satoshis
	}

	if inputSum <= outputSum {
		return errors.ErrOutputValueTooHigh
	}

	return nil
}

func findParentForInput(input *sdk.TransactionInput, parentTxs []*beef.TxData) *beef.TxData {
	parentID := input.SourceTXID.String()

	for _, ptx := range parentTxs {
		if ptx.GetTxID() == parentID {
			return ptx
		}
	}

	return nil
}
