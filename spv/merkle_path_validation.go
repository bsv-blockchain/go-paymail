package spv

import (
	"context"

	"github.com/bsv-blockchain/go-paymail/beef"
)

func verifyMerkleRoots(ctx context.Context, dBeef *beef.DecodedBEEF, provider MerkleRootVerifier) error {
	if err := ensureAncestorsArePresentInBump(dBeef.GetLatestTx(), dBeef); err != nil {
		return err
	}

	verifyReq, err := getMerkleRootsVerificationRequests(dBeef.BUMPs)
	if err != nil {
		return err
	}

	if err = provider.VerifyMerkleRoots(ctx, verifyReq); err != nil {
		return err
	}

	return nil
}

func getMerkleRootsVerificationRequests(bumps beef.BUMPs) ([]*MerkleRootConfirmationRequestItem, error) {
	reqItems := make([]*MerkleRootConfirmationRequestItem, 0, len(bumps))

	for _, bump := range bumps {
		merkleRoot, err := bump.CalculateMerkleRoot()
		if err != nil {
			return nil, err
		}

		req := MerkleRootConfirmationRequestItem{
			BlockHeight: bump.BlockHeight,
			MerkleRoot:  merkleRoot,
		}
		reqItems = append(reqItems, &req)
	}

	return reqItems, nil
}
