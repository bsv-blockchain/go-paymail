package beef

import (
	"errors"
)

var (
	// ErrBumpMerkleRootMismatch is returned when merkle roots don't match for the same block
	ErrBumpMerkleRootMismatch = errors.New("different merkle roots for the same block")
	// ErrBumpLeafNotComputable is returned when leaf cannot be computed from children at base level
	ErrBumpLeafNotComputable = errors.New("cannot compute leaf from children at base level")
	// ErrBumpChildNotFound is returned when child cannot be found
	ErrBumpChildNotFound = errors.New("could not find child")
	// ErrBumpPairNotFound is returned when pair cannot be found
	ErrBumpPairNotFound = errors.New("could not find pair")
)

// BUMPs represents a slice of BUMPs - BSV Unified Merkle Paths
type BUMPs []*BUMP

// BUMP is a struct that represents a whole BUMP format
type BUMP struct {
	BlockHeight uint64       `json:"blockHeight"`
	Path        [][]BUMPLeaf `json:"path"`
}

// BUMPLeaf represents each BUMP path element
type BUMPLeaf struct {
	Hash      string `json:"hash"`
	TxId      bool   `json:"txId"`
	Duplicate bool   `json:"duplicate"`
	Offset    uint64 `json:"offset"`
}

// Flags which are used to determine the type of BUMPLeaf
const (
	dataFlag byte = iota
	duplicateFlag
	txIDFlag
)

// CalculateMerkleRoot will calculate the merkle root for the BUMP
func (b BUMP) CalculateMerkleRoot() (string, error) {
	merkleRoot := ""

	for _, bumpPathElement := range b.Path[0] {
		if bumpPathElement.TxId {
			calcMerkleRoot, err := calculateMerkleRoot(bumpPathElement, b)
			if err != nil {
				return "", err
			}

			if merkleRoot == "" {
				merkleRoot = calcMerkleRoot
				continue
			}

			if calcMerkleRoot != merkleRoot {
				return "", ErrBumpMerkleRootMismatch
			}
		}
	}
	return merkleRoot, nil
}

// calculateMerkleRoots will calculate one merkle root for tx in the BUMPLeaf
func calculateMerkleRoot(baseLeaf BUMPLeaf, bump BUMP) (string, error) {
	calculatedHash := baseLeaf.Hash
	offset := baseLeaf.Offset

	for i := 0; i < len(bump.Path); i++ {
		bLevel := bump.Path[i]
		var previousBLevel []BUMPLeaf

		if i-1 >= 0 {
			previousBLevel = bump.Path[i-1]
		}
		newOffset := getOffsetPair(offset)
		leafInPair := findLeafByOffset(newOffset, bLevel)
		if leafInPair == nil {
			if previousBLevel == nil {
				return "", ErrBumpLeafNotComputable
			}

			var err error
			leafInPair, err = calculateFromChildren(newOffset, previousBLevel)
			if err != nil {
				return "", err
			}
		}

		leftNode, rightNode := prepareNodes(baseLeaf, offset, *leafInPair, newOffset)
		str, err := merkleTreeParentStr(leftNode, rightNode)
		if err != nil {
			return "", err
		}
		calculatedHash = str
		offset = offset / 2
		baseLeaf = BUMPLeaf{
			Hash:   calculatedHash,
			Offset: offset,
		}
	}

	return calculatedHash, nil
}

func getOffsetPair(offset uint64) uint64 {
	if offset%2 == 0 {
		return offset + 1
	}
	return offset - 1
}

func findLeafByOffset(offset uint64, bumpLeaves []BUMPLeaf) *BUMPLeaf {
	for _, bumpTx := range bumpLeaves {
		if bumpTx.Offset == offset {
			return &bumpTx
		}
	}
	return nil
}

func calculateFromChildren(offset uint64, bumpLeaves []BUMPLeaf) (*BUMPLeaf, error) {
	offsetChild := offset * 2
	offsetChildPair := offsetChild + 1
	leaf := findLeafByOffset(offsetChild, bumpLeaves)
	if leaf == nil {
		return nil, ErrBumpChildNotFound
	}

	leafInPair := findLeafByOffset(offsetChildPair, bumpLeaves)
	if leafInPair == nil {
		return nil, ErrBumpChildNotFound
	}

	leftNode, rightNode := prepareNodes(*leaf, offset, *leafInPair, offsetChildPair)
	str, err := merkleTreeParentStr(leftNode, rightNode)
	if err != nil {
		return nil, ErrBumpPairNotFound
	}

	return &BUMPLeaf{
		Hash:   str,
		Offset: offset,
	}, nil
}

func prepareNodes(baseLeaf BUMPLeaf, offset uint64, leafInPair BUMPLeaf, newOffset uint64) (string, string) {
	var baseLeafHash, pairLeafHash string

	if baseLeaf.Duplicate {
		baseLeafHash = leafInPair.Hash
	} else {
		baseLeafHash = baseLeaf.Hash
	}

	if leafInPair.Duplicate {
		pairLeafHash = baseLeaf.Hash
	} else {
		pairLeafHash = leafInPair.Hash
	}

	if newOffset > offset {
		return baseLeafHash, pairLeafHash
	}
	return pairLeafHash, baseLeafHash
}
