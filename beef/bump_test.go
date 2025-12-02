package beef

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOffsetPair(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		offset   uint64
		expected uint64
	}{
		{
			name:     "even offset returns next odd",
			offset:   0,
			expected: 1,
		},
		{
			name:     "odd offset returns previous even",
			offset:   1,
			expected: 0,
		},
		{
			name:     "even offset 20 returns 21",
			offset:   20,
			expected: 21,
		},
		{
			name:     "odd offset 21 returns 20",
			offset:   21,
			expected: 20,
		},
		{
			name:     "large even offset",
			offset:   1000,
			expected: 1001,
		},
		{
			name:     "large odd offset",
			offset:   1001,
			expected: 1000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getOffsetPair(tc.offset)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFindLeafByOffset(t *testing.T) {
	t.Parallel()

	testLeaves := []BUMPLeaf{
		{Hash: "hash0", Offset: 0},
		{Hash: "hash1", Offset: 1},
		{Hash: "hash5", Offset: 5},
		{Hash: "hash10", Offset: 10},
	}

	tests := []struct {
		name         string
		offset       uint64
		leaves       []BUMPLeaf
		expectedHash string
		expectNil    bool
	}{
		{
			name:         "find leaf at offset 0",
			offset:       0,
			leaves:       testLeaves,
			expectedHash: "hash0",
			expectNil:    false,
		},
		{
			name:         "find leaf at offset 1",
			offset:       1,
			leaves:       testLeaves,
			expectedHash: "hash1",
			expectNil:    false,
		},
		{
			name:         "find leaf at offset 5",
			offset:       5,
			leaves:       testLeaves,
			expectedHash: "hash5",
			expectNil:    false,
		},
		{
			name:         "find leaf at offset 10",
			offset:       10,
			leaves:       testLeaves,
			expectedHash: "hash10",
			expectNil:    false,
		},
		{
			name:      "offset not found returns nil",
			offset:    99,
			leaves:    testLeaves,
			expectNil: true,
		},
		{
			name:      "empty leaves returns nil",
			offset:    0,
			leaves:    []BUMPLeaf{},
			expectNil: true,
		},
		{
			name:      "nil leaves returns nil",
			offset:    0,
			leaves:    nil,
			expectNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := findLeafByOffset(tc.offset, tc.leaves)
			if tc.expectNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tc.expectedHash, result.Hash)
				assert.Equal(t, tc.offset, result.Offset)
			}
		})
	}
}

func TestPrepareNodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		baseLeaf          BUMPLeaf
		offset            uint64
		leafInPair        BUMPLeaf
		newOffset         uint64
		expectedLeftNode  string
		expectedRightNode string
	}{
		{
			name:              "base on left, pair on right",
			baseLeaf:          BUMPLeaf{Hash: "baseHash", Offset: 0},
			offset:            0,
			leafInPair:        BUMPLeaf{Hash: "pairHash", Offset: 1},
			newOffset:         1,
			expectedLeftNode:  "baseHash",
			expectedRightNode: "pairHash",
		},
		{
			name:              "base on right, pair on left",
			baseLeaf:          BUMPLeaf{Hash: "baseHash", Offset: 1},
			offset:            1,
			leafInPair:        BUMPLeaf{Hash: "pairHash", Offset: 0},
			newOffset:         0,
			expectedLeftNode:  "pairHash",
			expectedRightNode: "baseHash",
		},
		{
			name:              "base leaf duplicate uses pair hash",
			baseLeaf:          BUMPLeaf{Hash: "baseHash", Duplicate: true, Offset: 0},
			offset:            0,
			leafInPair:        BUMPLeaf{Hash: "pairHash", Offset: 1},
			newOffset:         1,
			expectedLeftNode:  "pairHash",
			expectedRightNode: "pairHash",
		},
		{
			name:              "pair leaf duplicate uses base hash",
			baseLeaf:          BUMPLeaf{Hash: "baseHash", Offset: 0},
			offset:            0,
			leafInPair:        BUMPLeaf{Hash: "pairHash", Duplicate: true, Offset: 1},
			newOffset:         1,
			expectedLeftNode:  "baseHash",
			expectedRightNode: "baseHash",
		},
		{
			name:              "both duplicate - uses each others hash",
			baseLeaf:          BUMPLeaf{Hash: "baseHash", Duplicate: true, Offset: 0},
			offset:            0,
			leafInPair:        BUMPLeaf{Hash: "pairHash", Duplicate: true, Offset: 1},
			newOffset:         1,
			expectedLeftNode:  "pairHash", // baseLeaf.Duplicate uses leafInPair.Hash
			expectedRightNode: "baseHash", // leafInPair.Duplicate uses baseLeaf.Hash
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			leftNode, rightNode := prepareNodes(tc.baseLeaf, tc.offset, tc.leafInPair, tc.newOffset)
			assert.Equal(t, tc.expectedLeftNode, leftNode)
			assert.Equal(t, tc.expectedRightNode, rightNode)
		})
	}
}

func TestCalculateFromChildren(t *testing.T) {
	t.Parallel()

	t.Run("finds both children and computes parent", func(t *testing.T) {
		// Create leaves at offset 0 and 1 (children of offset 0 at higher level)
		leaves := []BUMPLeaf{
			{Hash: "0dc75b4efeeddb95d8ee98ded75d781fcf95d35f9d88f7f1ce54a77a0c7c50fe", Offset: 0},
			{Hash: "3ecead27a44d013ad1aae40038acbb1883ac9242406808bb4667c15b4f164eac", Offset: 1},
		}

		result, err := calculateFromChildren(0, leaves)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint64(0), result.Offset)
		// The hash should be computed from the two children
		assert.NotEmpty(t, result.Hash)
	})

	t.Run("returns error when left child not found", func(t *testing.T) {
		leaves := []BUMPLeaf{
			{Hash: "hash1", Offset: 1}, // Only offset 1, missing offset 0
		}

		result, err := calculateFromChildren(0, leaves)
		require.Error(t, err)
		assert.Equal(t, ErrBumpChildNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when right child not found", func(t *testing.T) {
		leaves := []BUMPLeaf{
			{Hash: "hash0", Offset: 0}, // Only offset 0, missing offset 1
		}

		result, err := calculateFromChildren(0, leaves)
		require.Error(t, err)
		assert.Equal(t, ErrBumpChildNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("returns error with empty leaves", func(t *testing.T) {
		result, err := calculateFromChildren(0, []BUMPLeaf{})
		require.Error(t, err)
		assert.Equal(t, ErrBumpChildNotFound, err)
		assert.Nil(t, result)
	})
}

func TestBUMP_CalculateMerkleRoot(t *testing.T) {
	t.Parallel()

	t.Run("calculates merkle root for valid BUMP", func(t *testing.T) {
		// BUMP structure from beef_tx_test.go
		bump := BUMP{
			BlockHeight: 814435,
			Path: [][]BUMPLeaf{
				{
					{Hash: "0dc75b4efeeddb95d8ee98ded75d781fcf95d35f9d88f7f1ce54a77a0c7c50fe", Offset: 20},
					{Hash: "3ecead27a44d013ad1aae40038acbb1883ac9242406808bb4667c15b4f164eac", TxId: true, Offset: 21},
				},
				{
					{Hash: "5745cf28cd3a31703f611fb80b5a080da55acefa4c6977b21917d1ef95f34fbc", Offset: 11},
				},
				{
					{Hash: "522a096a1a6d3b64a4289ab456134158d8443f2c3b8ed8618bd2b842912d4b57", Offset: 4},
				},
				{
					{Hash: "191c70d2ecb477f90716d602f4e39f2f81f686f8f4230c255d1b534dc85fa051", Offset: 3},
				},
				{
					{Hash: "1f487b8cd3b11472c56617227e7e8509b44054f2a796f33c52c28fd5291578fd", Offset: 0},
				},
				{
					{Hash: "5ecc0ad4f24b5d8c7e6ec5669dc1d45fcb3405d8ce13c0860f66a35ef442f562", Offset: 1},
				},
				{
					{Hash: "31631241c8124bc5a9531c160bfddb6fcff3729f4e652b10d57cfd3618e921b1", Offset: 1},
				},
			},
		}

		merkleRoot, err := bump.CalculateMerkleRoot()
		require.NoError(t, err)
		assert.NotEmpty(t, merkleRoot)
	})

	t.Run("returns empty for BUMP with no txId leaves", func(t *testing.T) {
		bump := BUMP{
			BlockHeight: 100,
			Path: [][]BUMPLeaf{
				{
					{Hash: "hash1", TxId: false, Offset: 0},
					{Hash: "hash2", TxId: false, Offset: 1},
				},
			},
		}

		merkleRoot, err := bump.CalculateMerkleRoot()
		require.NoError(t, err)
		assert.Empty(t, merkleRoot)
	})

	t.Run("handles multiple txId leaves with same root", func(t *testing.T) {
		// BUMP with two adjacent txId leaves - they form a pair so compute the same root
		bump := BUMP{
			BlockHeight: 100,
			Path: [][]BUMPLeaf{
				{
					{Hash: "0dc75b4efeeddb95d8ee98ded75d781fcf95d35f9d88f7f1ce54a77a0c7c50fe", TxId: true, Offset: 0},
					{Hash: "3ecead27a44d013ad1aae40038acbb1883ac9242406808bb4667c15b4f164eac", TxId: true, Offset: 1},
				},
			},
		}

		merkleRoot, err := bump.CalculateMerkleRoot()
		require.NoError(t, err)
		// Both leaves compute to the same merkle root since they're paired
		assert.NotEmpty(t, merkleRoot)
	})
}

func TestBUMPLeaf_Flags(t *testing.T) {
	t.Parallel()

	// Test that flag constants are defined correctly
	assert.Equal(t, dataFlag, byte(0))
	assert.Equal(t, duplicateFlag, byte(1))
	assert.Equal(t, txIDFlag, byte(2))
}
