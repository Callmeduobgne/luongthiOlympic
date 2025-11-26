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
	"testing"
)

func TestGenerateMerkleTree(t *testing.T) {
	tests := []struct {
		name    string
		txIds   []string
		wantErr bool
	}{
		{
			name:    "Single transaction",
			txIds:   []string{"tx1"},
			wantErr: false,
		},
		{
			name:    "Two transactions",
			txIds:   []string{"tx1", "tx2"},
			wantErr: false,
		},
		{
			name:    "Four transactions (perfect binary tree)",
			txIds:   []string{"tx1", "tx2", "tx3", "tx4"},
			wantErr: false,
		},
		{
			name:    "Odd number of transactions",
			txIds:   []string{"tx1", "tx2", "tx3"},
			wantErr: false,
		},
		{
			name:    "Empty transaction list",
			txIds:   []string{},
			wantErr: true,
		},
		{
			name:    "Transaction with empty string",
			txIds:   []string{"tx1", "", "tx3"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := GenerateMerkleTree(tt.txIds)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMerkleTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tree == nil {
					t.Error("GenerateMerkleTree() returned nil tree")
					return
				}
				if tree.Root == nil {
					t.Error("GenerateMerkleTree() root is nil")
					return
				}
				if tree.TxCount != len(tt.txIds) {
					t.Errorf("GenerateMerkleTree() TxCount = %d, want %d", tree.TxCount, len(tt.txIds))
				}
				if len(tree.Leaves) != len(tt.txIds) {
					t.Errorf("GenerateMerkleTree() Leaves count = %d, want %d", len(tree.Leaves), len(tt.txIds))
				}
			}
		})
	}
}

func TestGenerateMerkleProof(t *testing.T) {
	txIds := []string{"tx1", "tx2", "tx3", "tx4"}
	tree, err := GenerateMerkleTree(txIds)
	if err != nil {
		t.Fatalf("Failed to create Merkle tree: %v", err)
	}

	tests := []struct {
		name    string
		txId    string
		wantErr bool
	}{
		{
			name:    "Proof for first transaction",
			txId:    "tx1",
			wantErr: false,
		},
		{
			name:    "Proof for last transaction",
			txId:    "tx4",
			wantErr: false,
		},
		{
			name:    "Proof for middle transaction",
			txId:    "tx2",
			wantErr: false,
		},
		{
			name:    "Proof for non-existent transaction",
			txId:    "tx999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proof, err := GenerateMerkleProof(tree, tt.txId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMerkleProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(proof) == 0 {
					t.Error("GenerateMerkleProof() returned empty proof")
				}
				// For a tree with 4 transactions, proof should have 2 hashes (tree depth - 1)
				expectedProofLength := tree.TreeDepth - 1
				if len(proof) != expectedProofLength {
					t.Errorf("GenerateMerkleProof() proof length = %d, want %d", len(proof), expectedProofLength)
				}
			}
		})
	}
}

func TestVerifyMerkleProof(t *testing.T) {
	txIds := []string{"tx1", "tx2", "tx3", "tx4"}
	tree, err := GenerateMerkleTree(txIds)
	if err != nil {
		t.Fatalf("Failed to create Merkle tree: %v", err)
	}

	// Generate valid proof
	proof, err := GenerateMerkleProof(tree, "tx1")
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	tests := []struct {
		name     string
		txId     string
		proof    []ProofStep
		rootHash string
		want     bool
	}{
		{
			name:     "Valid proof",
			txId:     "tx1",
			proof:    proof,
			rootHash: tree.Root.Hash,
			want:     true,
		},
		{
			name:     "Invalid proof - wrong transaction",
			txId:     "tx999",
			proof:    proof,
			rootHash: tree.Root.Hash,
			want:     false,
		},
		{
			name:     "Invalid proof - wrong root",
			txId:     "tx1",
			proof:    proof,
			rootHash: "wrong_root_hash",
			want:     false,
		},
		{
			name:     "Invalid proof - empty proof",
			txId:     "tx1",
			proof:    []ProofStep{},
			rootHash: tree.Root.Hash,
			want:     false,
		},
		{
			name:     "Invalid proof - tampered proof",
			txId:     "tx1",
			proof:    []ProofStep{{Hash: "tampered_hash1", Position: "left"}, {Hash: "tampered_hash2", Position: "right"}},
			rootHash: tree.Root.Hash,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyMerkleProof(tt.txId, tt.proof, tt.rootHash)
			if got != tt.want {
				t.Errorf("VerifyMerkleProof() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyMerkleProofAllTransactions(t *testing.T) {
	// Test that all transactions in the tree can be verified
	txIds := []string{"tx1", "tx2", "tx3", "tx4", "tx5", "tx6", "tx7", "tx8"}
	tree, err := GenerateMerkleTree(txIds)
	if err != nil {
		t.Fatalf("Failed to create Merkle tree: %v", err)
	}

	for _, txId := range txIds {
		proof, err := GenerateMerkleProof(tree, txId)
		if err != nil {
			t.Errorf("Failed to generate proof for %s: %v", txId, err)
			continue
		}

		if !VerifyMerkleProof(txId, proof, tree.Root.Hash) {
			t.Errorf("Failed to verify proof for %s", txId)
		}
	}
}

func TestComputeMerkleRoot(t *testing.T) {
	tests := []struct {
		name    string
		txIds   []string
		wantErr bool
	}{
		{
			name:    "Single transaction",
			txIds:   []string{"tx1"},
			wantErr: false,
		},
		{
			name:    "Multiple transactions",
			txIds:   []string{"tx1", "tx2", "tx3", "tx4"},
			wantErr: false,
		},
		{
			name:    "Empty list",
			txIds:   []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ComputeMerkleRoot(tt.txIds)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeMerkleRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if root == "" {
					t.Error("ComputeMerkleRoot() returned empty root")
				}
				// Verify root matches tree root
				tree, _ := GenerateMerkleTree(tt.txIds)
				if root != tree.Root.Hash {
					t.Errorf("ComputeMerkleRoot() = %s, want %s", root, tree.Root.Hash)
				}
			}
		})
	}
}

func TestVerifyTreeIntegrity(t *testing.T) {
	txIds := []string{"tx1", "tx2", "tx3", "tx4"}
	tree, err := GenerateMerkleTree(txIds)
	if err != nil {
		t.Fatalf("Failed to create Merkle tree: %v", err)
	}

	// Test valid tree
	if err := VerifyTreeIntegrity(tree); err != nil {
		t.Errorf("VerifyTreeIntegrity() failed for valid tree: %v", err)
	}

	// Test nil tree
	if err := VerifyTreeIntegrity(nil); err == nil {
		t.Error("VerifyTreeIntegrity() should fail for nil tree")
	}
}

func TestGetProofPath(t *testing.T) {
	txIds := []string{"tx1", "tx2", "tx3", "tx4"}
	tree, err := GenerateMerkleTree(txIds)
	if err != nil {
		t.Fatalf("Failed to create Merkle tree: %v", err)
	}

	proofPath, err := GetProofPath(tree, "tx1")
	if err != nil {
		t.Errorf("GetProofPath() error = %v", err)
		return
	}

	if len(proofPath) == 0 {
		t.Error("GetProofPath() returned empty proof path")
	}

	// Verify all steps have position
	for i, step := range proofPath {
		if step.Position != "left" && step.Position != "right" {
			t.Errorf("GetProofPath() step %d has invalid position: %s", i, step.Position)
		}
		if step.Hash == "" {
			t.Errorf("GetProofPath() step %d has empty hash", i)
		}
	}
}

// Benchmark tests
func BenchmarkGenerateMerkleTree(b *testing.B) {
	txIds := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		txIds[i] = "tx" + string(rune(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateMerkleTree(txIds)
	}
}

func BenchmarkGenerateMerkleProof(b *testing.B) {
	txIds := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		txIds[i] = "tx" + string(rune(i))
	}
	tree, _ := GenerateMerkleTree(txIds)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateMerkleProof(tree, "tx500")
	}
}

func BenchmarkVerifyMerkleProof(b *testing.B) {
	txIds := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		txIds[i] = "tx" + string(rune(i))
	}
	tree, _ := GenerateMerkleTree(txIds)
	proof, _ := GenerateMerkleProof(tree, "tx500")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyMerkleProof("tx500", proof, tree.Root.Hash)
	}
}
