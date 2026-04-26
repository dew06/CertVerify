package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type MerkleTree struct {
	Root   *Node
	Leaves []*Node
}

type Node struct {
	Hash  string
	Left  *Node
	Right *Node
}

// BuildTree creates Merkle tree from hashes
func BuildTree(hashes []string) (*MerkleTree, error) {
	if len(hashes) == 0 {
		return nil, errors.New("no hashes provided")
	}

	// Create leaf nodes
	var leaves []*Node
	for _, hash := range hashes {
		leaves = append(leaves, &Node{Hash: hash})
	}

	// Build tree bottom-up
	nodes := leaves
	for len(nodes) > 1 {
		var parents []*Node
		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			var right *Node
			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				right = left // Duplicate if odd number
			}

			parentHash := hashPair(left.Hash, right.Hash)
			parent := &Node{
				Hash:  parentHash,
				Left:  left,
				Right: right,
			}
			parents = append(parents, parent)
		}
		nodes = parents
	}

	return &MerkleTree{
		Root:   nodes[0],
		Leaves: leaves,
	}, nil
}

func hashPair(left, right string) string {
	combined := left + right
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}
