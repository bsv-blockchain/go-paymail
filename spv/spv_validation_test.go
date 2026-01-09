package spv

import (
	"testing"

	sdk "github.com/bsv-blockchain/go-sdk/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateLockTime(t *testing.T) {
	t.Parallel()

	t.Run("locktime zero is valid", func(t *testing.T) {
		tx := &sdk.Transaction{
			LockTime: 0,
			Inputs: []*sdk.TransactionInput{
				{SequenceNumber: 0xffffffff},
			},
		}

		err := validateLockTime(tx)
		assert.NoError(t, err)
	})

	t.Run("locktime non-zero with all max sequence is valid", func(t *testing.T) {
		tx := &sdk.Transaction{
			LockTime: 500000,
			Inputs: []*sdk.TransactionInput{
				{SequenceNumber: 0xffffffff},
				{SequenceNumber: 0xffffffff},
			},
		}

		err := validateLockTime(tx)
		assert.NoError(t, err)
	})

	t.Run("locktime non-zero with non-max sequence returns error", func(t *testing.T) {
		tx := &sdk.Transaction{
			LockTime: 500000,
			Inputs: []*sdk.TransactionInput{
				{SequenceNumber: 0xfffffffe}, // Not max
			},
		}

		err := validateLockTime(tx)
		require.Error(t, err)
	})

	t.Run("locktime non-zero with mixed sequence returns error", func(t *testing.T) {
		tx := &sdk.Transaction{
			LockTime: 500000,
			Inputs: []*sdk.TransactionInput{
				{SequenceNumber: 0xffffffff}, // Max
				{SequenceNumber: 0x00000001}, // Not max
			},
		}

		err := validateLockTime(tx)
		require.Error(t, err)
	})

	t.Run("empty inputs with non-zero locktime is valid", func(t *testing.T) {
		tx := &sdk.Transaction{
			LockTime: 500000,
			Inputs:   []*sdk.TransactionInput{},
		}

		// No inputs to check, so validation passes
		err := validateLockTime(tx)
		assert.NoError(t, err)
	})
}
