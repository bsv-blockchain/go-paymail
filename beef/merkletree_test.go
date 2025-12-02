package beef

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerkleTreeParentStr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		leftNode       string
		rightNode      string
		expectedParent string
		expectError    bool
	}{
		{
			name:           "valid merkle tree parent calculation",
			leftNode:       "0dc75b4efeeddb95d8ee98ded75d781fcf95d35f9d88f7f1ce54a77a0c7c50fe",
			rightNode:      "3ecead27a44d013ad1aae40038acbb1883ac9242406808bb4667c15b4f164eac",
			expectedParent: "31bdeeb51e5f4565dea35a41d3c4e463bd7c89ba129dbe2c8437683f24593559",
			expectError:    false,
		},
		{
			name:           "valid merkle parent with identical hashes",
			leftNode:       "5745cf28cd3a31703f611fb80b5a080da55acefa4c6977b21917d1ef95f34fbc",
			rightNode:      "5745cf28cd3a31703f611fb80b5a080da55acefa4c6977b21917d1ef95f34fbc",
			expectedParent: "300e15c914d9a5b290c516fafae4ed2cbcf7ae704ca8bfa1ba1e5ec2cfccbdd3",
			expectError:    false,
		},
		{
			name:        "invalid left node hex",
			leftNode:    "invalid_hex",
			rightNode:   "3ecead27a44d013ad1aae40038acbb1883ac9242406808bb4667c15b4f164eac",
			expectError: true,
		},
		{
			name:        "invalid right node hex",
			leftNode:    "0dc75b4efeeddb95d8ee98ded75d781fcf95d35f9d88f7f1ce54a77a0c7c50fe",
			rightNode:   "invalid_hex",
			expectError: true,
		},
		{
			name:        "empty left node",
			leftNode:    "",
			rightNode:   "3ecead27a44d013ad1aae40038acbb1883ac9242406808bb4667c15b4f164eac",
			expectError: false, // empty string decodes to empty bytes
		},
		{
			name:        "empty right node",
			leftNode:    "0dc75b4efeeddb95d8ee98ded75d781fcf95d35f9d88f7f1ce54a77a0c7c50fe",
			rightNode:   "",
			expectError: false, // empty string decodes to empty bytes
		},
		{
			name:        "odd length hex string",
			leftNode:    "abc",
			rightNode:   "3ecead27a44d013ad1aae40038acbb1883ac9242406808bb4667c15b4f164eac",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := merkleTreeParentStr(tc.leftNode, tc.rightNode)

			if tc.expectError {
				require.Error(t, err)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				if tc.expectedParent != "" {
					assert.Equal(t, tc.expectedParent, result)
				}
			}
		})
	}
}

func TestMerkleTreeParent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		leftNode       []byte
		rightNode      []byte
		expectedParent []byte
	}{
		{
			name:           "valid merkle tree parent calculation",
			leftNode:       hexToBytes(t, "0dc75b4efeeddb95d8ee98ded75d781fcf95d35f9d88f7f1ce54a77a0c7c50fe"),
			rightNode:      hexToBytes(t, "3ecead27a44d013ad1aae40038acbb1883ac9242406808bb4667c15b4f164eac"),
			expectedParent: hexToBytes(t, "31bdeeb51e5f4565dea35a41d3c4e463bd7c89ba129dbe2c8437683f24593559"),
		},
		{
			name:           "identical nodes",
			leftNode:       hexToBytes(t, "5745cf28cd3a31703f611fb80b5a080da55acefa4c6977b21917d1ef95f34fbc"),
			rightNode:      hexToBytes(t, "5745cf28cd3a31703f611fb80b5a080da55acefa4c6977b21917d1ef95f34fbc"),
			expectedParent: hexToBytes(t, "300e15c914d9a5b290c516fafae4ed2cbcf7ae704ca8bfa1ba1e5ec2cfccbdd3"),
		},
		{
			name:           "empty nodes",
			leftNode:       []byte{},
			rightNode:      []byte{},
			expectedParent: hexToBytes(t, "56944c5d3f98413ef45cf54545538103cc9f298e0575820ad3591376e2e0f65d"),
		},
		{
			name:           "nil nodes",
			leftNode:       nil,
			rightNode:      nil,
			expectedParent: hexToBytes(t, "56944c5d3f98413ef45cf54545538103cc9f298e0575820ad3591376e2e0f65d"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := merkleTreeParent(tc.leftNode, tc.rightNode)
			assert.Equal(t, tc.expectedParent, result)
		})
	}
}

func TestMerkleTreeParent_Consistency(t *testing.T) {
	t.Parallel()

	// Test that merkleTreeParentStr and merkleTreeParent produce consistent results
	leftHex := "0dc75b4efeeddb95d8ee98ded75d781fcf95d35f9d88f7f1ce54a77a0c7c50fe"
	rightHex := "3ecead27a44d013ad1aae40038acbb1883ac9242406808bb4667c15b4f164eac"

	// Get result from string function
	strResult, err := merkleTreeParentStr(leftHex, rightHex)
	require.NoError(t, err)

	// Get result from bytes function
	leftBytes := hexToBytes(t, leftHex)
	rightBytes := hexToBytes(t, rightHex)
	bytesResult := merkleTreeParent(leftBytes, rightBytes)

	// They should be equal
	assert.Equal(t, strResult, hex.EncodeToString(bytesResult))
}

// hexToBytes is a test helper to convert hex strings to bytes
func hexToBytes(t *testing.T, hexStr string) []byte {
	t.Helper()
	bytes, err := hex.DecodeString(hexStr)
	require.NoError(t, err)
	return bytes
}
