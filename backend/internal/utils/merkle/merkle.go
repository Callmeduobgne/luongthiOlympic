// Copyright 2024 IBN Network (ICTU Blockchain Network)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// MerkleTree represents a Merkle tree structure
type MerkleTree struct {
	Root         *MerkleNode
	Leaves       []*MerkleNode
	LeafMap      map[string]*MerkleNode // Map txId to leaf node for fast lookup
	TxCount      int
	TreeDepth    int
}

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Hash   string
	Left   *MerkleNode
	Right  *MerkleNode
	Parent *MerkleNode
	Data   string // Transaction ID (only for leaf nodes)
	IsLeaf bool
}

// GenerateMerkleTree creates a Merkle tree from transaction IDs
// This is the main entry point for creating a Merkle tree
func GenerateMerkleTree(txIds []string) (*MerkleTree, error) {
	if len(txIds) == 0 {
		return nil, errors.New("cannot create Merkle tree from empty transaction list")
	}

	// Create leaf nodes
	leaves := make([]*MerkleNode, len(txIds))
	leafMap := make(map[string]*MerkleNode)
	
	for i, txId := range txIds {
		if txId == "" {
			return nil, fmt.Errorf("transaction ID at index %d is empty", i)
		}
		
		hash := hashData(txId)
		leaf := &MerkleNode{
			Hash:   hash,
			Data:   txId,
			IsLeaf: true,
		}
		leaves[i] = leaf
		leafMap[txId] = leaf
	}

	// Build tree bottom-up
	root, depth := buildTree(leaves)

	return &MerkleTree{
		Root:      root,
		Leaves:    leaves,
		LeafMap:   leafMap,
		TxCount:   len(txIds),
		TreeDepth: depth,
	}, nil
}

// GenerateMerkleProof generates a Merkle proof for a specific transaction
// The proof is an array of sibling hashes with position information
func GenerateMerkleProof(tree *MerkleTree, txId string) ([]ProofStep, error) {
	if tree == nil {
		return nil, errors.New("merkle tree is nil")
	}

	// Find leaf node using map for O(1) lookup
	targetLeaf, exists := tree.LeafMap[txId]
	if !exists {
		return nil, fmt.Errorf("transaction %s not found in Merkle tree", txId)
	}

	// Generate proof path by collecting sibling hashes with positions
	proof := []ProofStep{}
	current := targetLeaf

	// Traverse up the tree until we reach the root
	for current.Parent != nil {
		parent := current.Parent
		
		// Get sibling node and determine position
		var sibling *MerkleNode
		var position string
		
		if parent.Left == current {
			sibling = parent.Right
			position = "right" // Sibling is on the right
		} else {
			sibling = parent.Left
			position = "left" // Sibling is on the left
		}

		// Add sibling hash and position to proof
		if sibling != nil {
			proof = append(proof, ProofStep{
				Hash:     sibling.Hash,
				Position: position,
			})
		}

		current = parent
	}

	return proof, nil
}

// VerifyMerkleProof verifies a Merkle proof against a root hash
// This is the core verification function that doesn't require blockchain access
func VerifyMerkleProof(txId string, proof []ProofStep, rootHash string) bool {
	if txId == "" || rootHash == "" {
		return false
	}

	// Start with the transaction hash
	currentHash := hashData(txId)

	// Apply each proof step to reconstruct the path to root
	for _, step := range proof {
		// Combine hashes based on position
		if step.Position == "left" {
			// Sibling is on the left, current is on the right
			currentHash = hashData(step.Hash + currentHash)
		} else {
			// Sibling is on the right, current is on the left
			currentHash = hashData(currentHash + step.Hash)
		}
	}

	// Compare computed root with provided root
	return currentHash == rootHash
}

// ComputeMerkleRoot computes the Merkle root from a list of transaction IDs
// This is a convenience function that creates a tree and returns just the root
func ComputeMerkleRoot(txIds []string) (string, error) {
	tree, err := GenerateMerkleTree(txIds)
	if err != nil {
		return "", err
	}
	return tree.Root.Hash, nil
}

// GetProofPath returns the proof path with position indicators (left/right)
// This is useful for debugging and visualization
func GetProofPath(tree *MerkleTree, txId string) ([]ProofStep, error) {
	if tree == nil {
		return nil, errors.New("merkle tree is nil")
	}

	targetLeaf, exists := tree.LeafMap[txId]
	if !exists {
		return nil, fmt.Errorf("transaction %s not found in Merkle tree", txId)
	}

	proofPath := []ProofStep{}
	current := targetLeaf

	for current.Parent != nil {
		parent := current.Parent
		
		var sibling *MerkleNode
		var position string
		
		if parent.Left == current {
			sibling = parent.Right
			position = "right"
		} else {
			sibling = parent.Left
			position = "left"
		}

		if sibling != nil {
			proofPath = append(proofPath, ProofStep{
				Hash:     sibling.Hash,
				Position: position,
			})
		}

		current = parent
	}

	return proofPath, nil
}

// ProofStep represents a step in the Merkle proof path
type ProofStep struct {
	Hash     string `json:"hash"`
	Position string `json:"position"` // "left" or "right"
}

// buildTree builds a Merkle tree from leaf nodes using bottom-up approach
// Returns the root node and tree depth
func buildTree(nodes []*MerkleNode) (*MerkleNode, int) {
	if len(nodes) == 1 {
		return nodes[0], 1
	}

	// Pair up nodes and create parent level
	var parents []*MerkleNode
	for i := 0; i < len(nodes); i += 2 {
		left := nodes[i]
		var right *MerkleNode
		
		if i+1 < len(nodes) {
			right = nodes[i+1]
		} else {
			// If odd number of nodes, duplicate the last one
			right = left
		}

		// Create parent node
		parentHash := hashData(left.Hash + right.Hash)
		parent := &MerkleNode{
			Hash:   parentHash,
			Left:   left,
			Right:  right,
			IsLeaf: false,
		}

		// Set parent references
		left.Parent = parent
		if right != left {
			right.Parent = parent
		}

		parents = append(parents, parent)
	}

	// Recursively build upper levels
	root, childDepth := buildTree(parents)
	return root, childDepth + 1
}

// hashData computes SHA-256 hash of data and returns hex string
func hashData(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// VerifyTreeIntegrity verifies the integrity of the entire Merkle tree
// This is useful for testing and debugging
func VerifyTreeIntegrity(tree *MerkleTree) error {
	if tree == nil {
		return errors.New("tree is nil")
	}

	if tree.Root == nil {
		return errors.New("tree root is nil")
	}

	// Verify all leaf nodes
	for txId, leaf := range tree.LeafMap {
		expectedHash := hashData(txId)
		if leaf.Hash != expectedHash {
			return fmt.Errorf("leaf hash mismatch for tx %s", txId)
		}
	}

	// Verify all parent nodes
	return verifyNodeIntegrity(tree.Root)
}

// verifyNodeIntegrity recursively verifies node hash integrity
func verifyNodeIntegrity(node *MerkleNode) error {
	if node == nil {
		return nil
	}

	// Leaf nodes are already verified
	if node.IsLeaf {
		return nil
	}

	// Verify parent hash
	if node.Left != nil && node.Right != nil {
		expectedHash := hashData(node.Left.Hash + node.Right.Hash)
		if node.Hash != expectedHash {
			return fmt.Errorf("parent hash mismatch: expected %s, got %s", expectedHash, node.Hash)
		}
	}

	// Recursively verify children
	if err := verifyNodeIntegrity(node.Left); err != nil {
		return err
	}
	if err := verifyNodeIntegrity(node.Right); err != nil {
		return err
	}

	return nil
}
