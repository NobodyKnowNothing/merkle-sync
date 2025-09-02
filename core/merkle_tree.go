package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Hash     string
	Left     *MerkleNode
	Right    *MerkleNode
	IsLeaf   bool
	Data     []byte // Only for leaf nodes
	BlockID  string // Only for leaf nodes
}

// MerkleTree represents the entire Merkle tree
type MerkleTree struct {
	Root     *MerkleNode
	Leaves   []*MerkleNode
	RootHash string
}

// DataBlock represents an encrypted data block
type DataBlock struct {
	ID            string
	EncryptedData []byte
	TableName     string
	Operation     string
	Timestamp     int64
	Metadata      map[string]string
}

// NewMerkleTree creates a new Merkle tree from a list of data blocks
func NewMerkleTree(blocks []DataBlock) (*MerkleTree, error) {
	if len(blocks) == 0 {
		return &MerkleTree{
			Root:     nil,
			Leaves:   []*MerkleNode{},
			RootHash: "",
		}, nil
	}

	// Create leaf nodes
	leaves := make([]*MerkleNode, len(blocks))
	for i, block := range blocks {
		leafHash := HashData(block.EncryptedData)
		leaves[i] = &MerkleNode{
			Hash:    leafHash,
			IsLeaf:  true,
			Data:    block.EncryptedData,
			BlockID: block.ID,
		}
	}

	// Build the tree
	root := buildTree(leaves)
	rootHash := ""
	if root != nil {
		rootHash = root.Hash
	}

	return &MerkleTree{
		Root:     root,
		Leaves:   leaves,
		RootHash: rootHash,
	}, nil
}

// buildTree recursively builds the Merkle tree from leaf nodes
func buildTree(nodes []*MerkleNode) *MerkleNode {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) == 1 {
		return nodes[0]
	}

	// If odd number of nodes, duplicate the last one
	if len(nodes)%2 == 1 {
		nodes = append(nodes, nodes[len(nodes)-1])
	}

	// Create parent nodes
	parents := make([]*MerkleNode, len(nodes)/2)
	for i := 0; i < len(nodes); i += 2 {
		left := nodes[i]
		right := nodes[i+1]
		
		// Create parent hash by concatenating left and right hashes
		parentHash := HashConcat(left.Hash, right.Hash)
		
		parents[i/2] = &MerkleNode{
			Hash:   parentHash,
			Left:   left,
			Right:  right,
			IsLeaf: false,
		}
	}

	return buildTree(parents)
}

// GenerateProof generates a Merkle proof for the given leaf hashes
func (mt *MerkleTree) GenerateProof(leafHashes []string) ([]ProofNode, error) {
	if mt.Root == nil {
		return nil, fmt.Errorf("empty tree")
	}

	// Find the leaf nodes for the given hashes
	leafNodes := make([]*MerkleNode, 0)
	for _, hash := range leafHashes {
		for _, leaf := range mt.Leaves {
			if leaf.Hash == hash {
				leafNodes = append(leafNodes, leaf)
				break
			}
		}
	}

	if len(leafNodes) == 0 {
		return nil, fmt.Errorf("no matching leaf nodes found")
	}

	// Generate proof path
	proofPath := make([]ProofNode, 0)
	visited := make(map[string]bool)

	for _, leaf := range leafNodes {
		path := mt.getProofPath(leaf, visited)
		proofPath = append(proofPath, path...)
	}

	return proofPath, nil
}

// getProofPath gets the proof path for a specific leaf node
func (mt *MerkleTree) getProofPath(leaf *MerkleNode, visited map[string]bool) []ProofNode {
	proofPath := make([]ProofNode, 0)
	current := leaf

	// Traverse up the tree to find the path to root
	for current != mt.Root {
		parent := mt.findParent(current)
		if parent == nil {
			break
		}

		// Determine if current is left or right child
		isLeft := parent.Left == current
		var sibling *MerkleNode
		if isLeft {
			sibling = parent.Right
		} else {
			sibling = parent.Left
		}

		// Add sibling to proof path if not already visited
		if sibling != nil && !visited[sibling.Hash] {
			proofPath = append(proofPath, ProofNode{
				Hash:    sibling.Hash,
				IsLeft:  !isLeft, // Sibling's position relative to parent
			})
			visited[sibling.Hash] = true
		}

		current = parent
	}

	return proofPath
}

// findParent finds the parent of a given node
func (mt *MerkleTree) findParent(target *MerkleNode) *MerkleNode {
	return mt.findParentRecursive(mt.Root, target)
}

// findParentRecursive recursively searches for the parent of a target node
func (mt *MerkleTree) findParentRecursive(current, target *MerkleNode) *MerkleNode {
	if current == nil || current.IsLeaf {
		return nil
	}

	if current.Left == target || current.Right == target {
		return current
	}

	// Search in left subtree
	if parent := mt.findParentRecursive(current.Left, target); parent != nil {
		return parent
	}

	// Search in right subtree
	return mt.findParentRecursive(current.Right, target)
}

// VerifyProof verifies a Merkle proof
func VerifyProof(rootHash string, leafHashes []string, proofPath []ProofNode) (bool, error) {
	if len(leafHashes) == 0 {
		return false, fmt.Errorf("no leaf hashes provided")
	}

	// Sort leaf hashes for consistent ordering
	sortedLeaves := make([]string, len(leafHashes))
	copy(sortedLeaves, leafHashes)
	sort.Strings(sortedLeaves)

	// Reconstruct the root hash
	computedRoot := reconstructRoot(sortedLeaves, proofPath)
	
	return computedRoot == rootHash, nil
}



// reconstructRoot reconstructs the root hash from leaf hashes and proof path
func reconstructRoot(leafHashes []string, proofPath []ProofNode) string {
	if len(leafHashes) == 0 {
		return ""
	}

	// For multiple leaf hashes, we need to reconstruct the entire tree
	if len(leafHashes) > 1 {
		return reconstructRootMultiple(leafHashes, proofPath)
	}

	// For single leaf hash, reconstruct the path to root
	currentHash := leafHashes[0]
	
	for _, proofNode := range proofPath {
		if proofNode.IsLeft {
			// Proof node is on the left, current hash is on the right
			currentHash = HashConcat(proofNode.Hash, currentHash)
		} else {
			// Proof node is on the right, current hash is on the left
			currentHash = HashConcat(currentHash, proofNode.Hash)
		}
	}
	
	return currentHash
}

// reconstructRootMultiple handles multiple leaf hashes by building the tree level by level
func reconstructRootMultiple(leafHashes []string, proofPath []ProofNode) string {
	// For multiple leaf hashes, we need to reconstruct the tree by following the proof path
	// The proof nodes represent the siblings needed to reconstruct the path to root
	
	// Create a map of all available hashes (leaves + proof nodes)
	availableHashes := make(map[string]bool)
	for _, hash := range leafHashes {
		availableHashes[hash] = true
	}
	for _, proofNode := range proofPath {
		availableHashes[proofNode.Hash] = true
	}
	
	// Start with leaf hashes
	currentLevel := make([]string, 0)
	for _, hash := range leafHashes {
		if availableHashes[hash] {
			currentLevel = append(currentLevel, hash)
		}
	}
	
	// Sort for consistent ordering
	sort.Strings(currentLevel)
	
	// Build the tree level by level
	for len(currentLevel) > 1 {
		nextLevel := make([]string, 0)
		
		// Process pairs
		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				// We have a pair, combine them
				combined := HashConcat(currentLevel[i], currentLevel[i+1])
				nextLevel = append(nextLevel, combined)
			} else {
				// Odd number of nodes, duplicate the last one
				combined := HashConcat(currentLevel[i], currentLevel[i])
				nextLevel = append(nextLevel, combined)
			}
		}
		
		currentLevel = nextLevel
	}
	
	if len(currentLevel) == 1 {
		return currentLevel[0]
	}
	
	return ""
}

// ProofNode represents a node in a Merkle proof
type ProofNode struct {
	Hash   string
	IsLeft bool
}

// HashData hashes the given data with a prefix to prevent second-preimage attacks
func HashData(data []byte) string {
	// Add prefix to distinguish leaf hashes from internal node hashes
	prefix := []byte("LEAF:")
	combined := append(prefix, data...)
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:])
}

// HashConcat hashes the concatenation of two hashes
func HashConcat(hash1, hash2 string) string {
	// Add prefix to distinguish internal node hashes
	prefix := []byte("INTERNAL:")
	combined := append(prefix, []byte(hash1)...)
	combined = append(combined, []byte(hash2)...)
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:])
}

// HashToString converts a byte slice to a hex string hash
func HashToString(data []byte) string {
	return hex.EncodeToString(data)
}

// StringToHash converts a hex string back to bytes
func StringToHash(hashStr string) ([]byte, error) {
	return hex.DecodeString(hashStr)
}

// DiffTrees compares two Merkle trees and returns differences
func DiffTrees(root1, root2 *MerkleTree) ([]DiffNode, error) {
	if root1 == nil && root2 == nil {
		return []DiffNode{}, nil
	}
	if root1 == nil || root2 == nil {
		return []DiffNode{}, fmt.Errorf("one tree is nil")
	}

	differences := make([]DiffNode, 0)
	diffNodes(root1.Root, root2.Root, &differences)
	
	return differences, nil
}

// diffNodes recursively compares two nodes and their subtrees
func diffNodes(node1, node2 *MerkleNode, differences *[]DiffNode) {
	if node1 == nil && node2 == nil {
		return
	}
	
	if node1 == nil || node2 == nil {
		// One node is nil, entire subtree is different
		if node1 != nil {
			*differences = append(*differences, DiffNode{
				Hash: node1.Hash,
				IsLeaf: node1.IsLeaf,
			})
		}
		if node2 != nil {
			*differences = append(*differences, DiffNode{
				Hash: node2.Hash,
				IsLeaf: node2.IsLeaf,
			})
		}
		return
	}
	
	if node1.Hash != node2.Hash {
		// Nodes are different
		if node1.IsLeaf && node2.IsLeaf {
			// Both are leaves, add both as differences
			*differences = append(*differences, DiffNode{
				Hash: node1.Hash,
				IsLeaf: true,
			})
			*differences = append(*differences, DiffNode{
				Hash: node2.Hash,
				IsLeaf: true,
			})
		} else {
			// Recurse into children
			diffNodes(node1.Left, node2.Left, differences)
			diffNodes(node1.Right, node2.Right, differences)
		}
	}
}

// DiffNode represents a difference between two trees
type DiffNode struct {
	Hash   string
	IsLeaf bool
}
