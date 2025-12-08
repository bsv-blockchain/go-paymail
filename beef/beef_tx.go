package beef

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"

	sdk "github.com/bitcoin-sv/go-sdk/transaction"
	util "github.com/bitcoin-sv/go-sdk/util"
)

const (
	BEEFMarkerPart1 = 0xBE
	BEEFMarkerPart2 = 0xEF
)

const (
	HasNoBump = 0x00
	HasBump   = 0x01
)

const (
	hashBytesCount    = 32
	markerBytesCount  = 2
	versionBytesCount = 2
	maxTreeHeight     = 64
)

var (
	// ErrBeefNoBytesForBump is returned when there are no bytes to decode BUMP
	ErrBeefNoBytesForBump = errors.New("cannot decode BUMP - no bytes provided")
	// ErrBeefNoLowestBump is returned when BEEF lacks BUMPs
	ErrBeefNoLowestBump = errors.New("invalid BEEF- lack of BUMPs")
	// ErrBeefInsufficientBytesBlockHeight is returned when there are insufficient bytes for BUMP blockHeight
	ErrBeefInsufficientBytesBlockHeight = errors.New("insufficient bytes to extract BUMP blockHeight")
	// ErrBeefInvalidTreeHeight is returned when treeHeight is invalid
	ErrBeefInvalidTreeHeight = errors.New("invalid BEEF - treeHeight cannot be grater than maxTreeHeight")
	// ErrBeefNoBytesForPaths is returned when there are no bytes for BUMP paths
	ErrBeefNoBytesForPaths = errors.New("cannot decode BUMP paths number of leaves from stream - no bytes provided")
	// ErrBeefInsufficientBytesHash is returned when there are insufficient bytes for hash
	ErrBeefInsufficientBytesHash = errors.New("insufficient bytes to extract hash of path")
	// ErrBeefInsufficientTransactions is returned when there are insufficient transactions
	ErrBeefInsufficientTransactions = errors.New("invalid BEEF- not enough transactions provided to decode BEEF")
	// ErrBeefInvalidHexStream is returned when BEEF hex stream is invalid
	ErrBeefInvalidHexStream = errors.New("invalid beef hex stream")
	// ErrBeefInvalidMarker is returned when BEEF marker is not found
	ErrBeefInvalidMarker = errors.New("invalid format of transaction, BEEF marker not found")
	// ErrBeefInvalidTreeHeightMax is returned when treeHeight is greater than 64
	ErrBeefInvalidTreeHeightMax = errors.New("invalid BEEF - treeHeight cannot be grater than 64")
	// ErrBeefInsufficientBytesOffset is returned when there are insufficient bytes to extract offset
	ErrBeefInsufficientBytesOffset = errors.New("insufficient bytes to extract offset")
	// ErrBeefInsufficientBytesFlag is returned when there are insufficient bytes to extract flag
	ErrBeefInsufficientBytesFlag = errors.New("insufficient bytes to extract flag")
	// ErrBeefInvalidFlag is returned when flag is invalid
	ErrBeefInvalidFlag = errors.New("invalid flag")
	// ErrBeefInvalidHasCMPFlag is returned when HasCMP flag is invalid
	ErrBeefInvalidHasCMPFlag = errors.New("invalid HasCMP flag for transaction")
	// ErrBeefIntegerOverflow is returned when an integer conversion would cause overflow
	ErrBeefIntegerOverflow = errors.New("integer value exceeds maximum safe conversion range")
)

type TxData struct {
	Transaction *sdk.Transaction `json:"transaction"`
	BumpIndex   *sdk.VarInt      `json:"bumpIndex"`

	txID string
}

func (td *TxData) Unmined() bool {
	return td.BumpIndex == nil
}

func (td *TxData) GetTxID() string {
	if len(td.txID) == 0 {
		td.txID = td.Transaction.TxID().String()
	}

	return td.txID
}

type DecodedBEEF struct {
	BUMPs        BUMPs     `json:"bumps"`
	Transactions []*TxData `json:"transactions"`
}

func DecodeBEEF(beefHex string) (*DecodedBEEF, error) {
	beefBytes, err := extractBytesWithoutVersionAndMarker(beefHex)
	if err != nil {
		return nil, err
	}

	bumps, remainingBytes, err := decodeBUMPs(beefBytes)
	if err != nil {
		return nil, err
	}

	transactions, err := decodeTransactionsWithPathIndexes(remainingBytes)
	if err != nil {
		return nil, err
	}

	return &DecodedBEEF{
		BUMPs:        bumps,
		Transactions: transactions,
	}, nil
}

func (d *DecodedBEEF) GetLatestTx() *sdk.Transaction {
	return d.Transactions[len(d.Transactions)-1].Transaction // get the last transaction as the processed transaction - it should be the last one because of khan's ordering
}

func decodeBUMPs(beefBytes []byte) ([]*BUMP, []byte, error) {
	if len(beefBytes) == 0 {
		return nil, nil, ErrBeefNoBytesForBump
	}

	nBump, bytesUsed := sdk.NewVarIntFromBytes(beefBytes)

	if nBump == 0 {
		return nil, nil, ErrBeefNoLowestBump
	}

	beefBytes = beefBytes[bytesUsed:]

	bumps := make([]*BUMP, 0, uint64(nBump))
	for i := uint64(0); i < uint64(nBump); i++ {
		if len(beefBytes) == 0 {
			return nil, nil, ErrBeefInsufficientBytesBlockHeight
		}
		blockHeight, bytesUsed := sdk.NewVarIntFromBytes(beefBytes)
		beefBytes = beefBytes[bytesUsed:]

		treeHeight := beefBytes[0]
		if int(treeHeight) > maxTreeHeight {
			return nil, nil, fmt.Errorf("treeHeight: %d: %w", treeHeight, ErrBeefInvalidTreeHeight)
		}
		beefBytes = beefBytes[1:]

		bumpPaths, remainingBytes, err := decodeBUMPPathsFromStream(int(treeHeight), beefBytes)
		if err != nil {
			return nil, nil, err
		}
		beefBytes = remainingBytes

		bump := &BUMP{
			BlockHeight: uint64(blockHeight),
			Path:        bumpPaths,
		}

		bumps = append(bumps, bump)
	}

	return bumps, beefBytes, nil
}

func decodeBUMPPathsFromStream(treeHeight int, hexBytes []byte) ([][]BUMPLeaf, []byte, error) {
	bumpPaths := make([][]BUMPLeaf, 0)

	for i := 0; i < treeHeight; i++ {
		if len(hexBytes) == 0 {
			return nil, nil, ErrBeefNoBytesForPaths
		}
		nLeaves, bytesUsed := sdk.NewVarIntFromBytes(hexBytes)
		hexBytes = hexBytes[bytesUsed:]
		bumpPath, remainingBytes, err := decodeBUMPLevel(nLeaves, hexBytes)
		if err != nil {
			return nil, nil, err
		}
		hexBytes = remainingBytes
		bumpPaths = append(bumpPaths, bumpPath)
	}

	return bumpPaths, hexBytes, nil
}

func decodeBUMPLevel(nLeaves sdk.VarInt, hexBytes []byte) ([]BUMPLeaf, []byte, error) {
	// Check for integer overflow before converting uint64 to int
	if nLeaves > math.MaxInt {
		return nil, nil, fmt.Errorf("number of leaves %d: %w", nLeaves, ErrBeefIntegerOverflow)
	}
	nLeavesInt := int(nLeaves) // #nosec G115 - overflow checked above

	bumpPath := make([]BUMPLeaf, 0)
	for i := 0; i < nLeavesInt; i++ {
		if len(hexBytes) == 0 {
			return nil, nil, fmt.Errorf("leaf %d of %d: %w", i, nLeavesInt, ErrBeefInsufficientBytesOffset)
		}

		offset, bytesUsed := sdk.NewVarIntFromBytes(hexBytes)
		hexBytes = hexBytes[bytesUsed:]

		if len(hexBytes) == 0 {
			return nil, nil, fmt.Errorf("leaf %d of %d: %w", i, nLeavesInt, ErrBeefInsufficientBytesFlag)
		}

		flag := hexBytes[0]
		hexBytes = hexBytes[1:]

		if flag != dataFlag && flag != duplicateFlag && flag != txIDFlag {
			return nil, nil, fmt.Errorf("flag %d for leaf %d of %d: %w", flag, i, nLeavesInt, ErrBeefInvalidFlag)
		}

		if flag == duplicateFlag {
			bumpLeaf := BUMPLeaf{
				Offset:    uint64(offset),
				Duplicate: true,
			}
			bumpPath = append(bumpPath, bumpLeaf)
			continue
		}

		if len(hexBytes) < hashBytesCount {
			return nil, nil, ErrBeefInsufficientBytesHash
		}

		hash := hex.EncodeToString(util.ReverseBytes(hexBytes[:hashBytesCount]))
		hexBytes = hexBytes[hashBytesCount:]

		bumpLeaf := BUMPLeaf{
			Hash:   hash,
			Offset: uint64(offset),
		}
		if flag == txIDFlag {
			bumpLeaf.TxId = true
		}
		bumpPath = append(bumpPath, bumpLeaf)
	}

	return bumpPath, hexBytes, nil
}

func decodeTransactionsWithPathIndexes(bytes []byte) ([]*TxData, error) {
	nTransactions, offset := sdk.NewVarIntFromBytes(bytes)

	if nTransactions < 2 {
		return nil, ErrBeefInsufficientTransactions
	}

	// Check for integer overflow before converting uint64 to int
	if nTransactions > math.MaxInt {
		return nil, fmt.Errorf("number of transactions %d: %w", nTransactions, ErrBeefIntegerOverflow)
	}
	nTransactionsInt := int(nTransactions) // #nosec G115 - overflow checked above

	bytes = bytes[offset:]

	transactions := make([]*TxData, 0, nTransactionsInt)

	for i := 0; i < nTransactionsInt; i++ {
		tx, offset, err := sdk.NewTransactionFromStream(bytes)
		if err != nil {
			return nil, err
		}
		bytes = bytes[offset:]

		var pathIndex *sdk.VarInt

		switch bytes[0] {
		case HasBump:
			value, offset := sdk.NewVarIntFromBytes(bytes[1:])
			pathIndex = &value
			bytes = bytes[1+offset:]
		case HasNoBump:
			bytes = bytes[1:]
		default:
			return nil, fmt.Errorf("transaction at index %d: %w", i, ErrBeefInvalidHasCMPFlag)
		}

		transactions = append(transactions, &TxData{
			Transaction: tx,
			BumpIndex:   pathIndex,
		})
	}

	return transactions, nil
}

func extractBytesWithoutVersionAndMarker(hexStream string) ([]byte, error) {
	bytes, err := hex.DecodeString(hexStream)
	if err != nil {
		return nil, ErrBeefInvalidHexStream
	}
	if len(bytes) < 4 {
		return nil, ErrBeefInvalidHexStream
	}

	// removes version bytes
	bytes = bytes[versionBytesCount:]
	err = validateMarker(bytes)
	if err != nil {
		return nil, err
	}

	// removes marker bytes
	bytes = bytes[markerBytesCount:]

	return bytes, nil
}

func validateMarker(bytes []byte) error {
	if bytes[0] != BEEFMarkerPart1 || bytes[1] != BEEFMarkerPart2 {
		return ErrBeefInvalidMarker
	}

	return nil
}
