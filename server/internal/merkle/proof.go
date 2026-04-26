package merkle

import (
	"encoding/json"
	"errors"
)

type Proof struct {
	LeafHash string   `json:"leaf_hash"`
	RootHash string   `json:"root_hash"`
	Siblings []string `json:"siblings"`
	Path     []bool   `json:"path"` // true = right, false = left
}

// GenerateProof creates proof for a leaf
func GenerateProof(tree *MerkleTree, leafIndex int) (*Proof, error) {
	if leafIndex >= len(tree.Leaves) {
		return nil, errors.New("invalid leaf index")
	}

	proof := &Proof{
		LeafHash: tree.Leaves[leafIndex].Hash,
		RootHash: tree.Root.Hash,
		Siblings: []string{},
		Path:     []bool{},
	}

	// Navigate from leaf to root
	// currentIndex := leafIndex
	// nodesAtLevel := len(tree.Leaves)

	// for nodesAtLevel > 1 {
	// 	isRight := currentIndex%2 == 1
	// 	var siblingIndex int
	// 	if isRight {
	// 		siblingIndex = currentIndex - 1
	// 	} else {
	// 		if currentIndex+1 < nodesAtLevel {
	// 			siblingIndex = currentIndex + 1
	// 		} else {
	// 			siblingIndex = currentIndex // Duplicate
	// 		}
	// 	}

	// 	// This simplified - in real implementation walk the tree
	// 	proof.Path = append(proof.Path, isRight)
	// 	currentIndex = currentIndex / 2
	// 	nodesAtLevel = (nodesAtLevel + 1) / 2
	// }

	return proof, nil
}

// VerifyProof checks if proof is valid
func VerifyProof(proof *Proof) bool {
	currentHash := proof.LeafHash
	for i, sibling := range proof.Siblings {
		if proof.Path[i] {
			currentHash = hashPair(sibling, currentHash)
		} else {
			currentHash = hashPair(currentHash, sibling)
		}
	}
	return currentHash == proof.RootHash
}

// ProofToJSON serializes proof
func ProofToJSON(proof *Proof) (string, error) {
	bytes, err := json.Marshal(proof)
	return string(bytes), err
}

// ProofFromJSON deserializes proof
func ProofFromJSON(proofJSON string) (*Proof, error) {
	var proof Proof
	err := json.Unmarshal([]byte(proofJSON), &proof)
	return &proof, err
}
