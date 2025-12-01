package spv

import (
	interpreter "github.com/bitcoin-sv/go-sdk/script/interpreter"
	sdk "github.com/bitcoin-sv/go-sdk/transaction"

	"github.com/bsv-blockchain/go-paymail/beef"
	"github.com/bsv-blockchain/go-paymail/errors"
)

func validateScripts(tx *sdk.Transaction, inputTxs []*beef.TxData) error {
	for i, input := range tx.Inputs {
		inputParentTx := findParentForInput(input, inputTxs)
		if inputParentTx == nil {
			return errors.ErrNoMatchingTransactionsForInput
		}

		err := verifyScripts(tx, inputParentTx.Transaction, i)
		if err != nil {
			return errors.ErrInvalidScript
		}
	}

	return nil
}

// Verify locking and unlocking scripts pair
func verifyScripts(tx, prevTx *sdk.Transaction, inputIdx int) error {
	input := tx.InputIdx(inputIdx)
	prevOutput := prevTx.OutputIdx(int(input.SourceTxOutIndex))

	err := interpreter.NewEngine().Execute(
		interpreter.WithTx(tx, inputIdx, prevOutput),
		interpreter.WithForkID(),
		interpreter.WithAfterGenesis(),
	)

	return err
}
