package bedtree

import (
	"math"
)

type BPlusTree struct {
	branchFactor int
	root         *bPlusTreeNode
	compare      func(string, string) int
}

type bPlusTreeNode struct {
	parent         *bPlusTreeNode
	parentChildIdx int
	splits         []string         // size m
	children       []*bPlusTreeNode // size m + 1 (internal only)
	data           [][]interface{}  // size m x n (leaf only)
}

func intMax(a int, b int) int {
	return int(math.Max(float64(a), float64(b)))
}

func intMin(a int, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

func intAbs(a int) int {
	return int(math.Abs(float64(a)))
}

func New(b int, compare func(string, string) int) *BPlusTree {
	if b < 2 {
		return nil
	}

	tree := new(BPlusTree)
	tree.branchFactor = b
	tree.compare = compare
	tree.root = tree.createTreeNode()
	return tree
}

func (tree *BPlusTree) Insert(key string) {
	nodeToInsert := tree.recFindNode(key, tree.root)
	tree.recInsert(key, nodeToInsert, nil, nil)
}

func (tree *BPlusTree) Put(key string, value interface{}) {
	nodeToInsert := tree.recFindNode(key, tree.root)
	tree.recInsert(key, nodeToInsert, value, nil)
}

func (tree *BPlusTree) createTreeNode() *bPlusTreeNode {
	node := new(bPlusTreeNode)
	node.splits = make([]string, 0, tree.branchFactor-1)
	node.children = make([]*bPlusTreeNode, 0, tree.branchFactor)
	return node
}

func (tree *BPlusTree) addToParentNode(parent *bPlusTreeNode, child *bPlusTreeNode) {
	if child != nil {
		nodeWithSplitValue := child
		for len(nodeWithSplitValue.splits) == 0 {
			nodeWithSplitValue = nodeWithSplitValue.children[0]
		}

		splitValue := nodeWithSplitValue.splits[0]
		if parent == nil {
			newRoot := tree.createTreeNode()
			newRoot.children = append(newRoot.children, tree.root)
			newRoot.children = append(newRoot.children, child)
			newRoot.splits = append(newRoot.splits, splitValue)
			tree.root.parent = newRoot
			tree.root.parentChildIdx = 0
			child.parent = newRoot
			child.parentChildIdx = 1
			tree.root = newRoot
		} else {
			tree.recInsert(splitValue, parent, nil, child)
		}
	}
}

func (tree *BPlusTree) splitIfNecessary(parent *bPlusTreeNode, q string) (*bPlusTreeNode, bool) {

	shouldSplit := false
	if len(parent.splits) == tree.branchFactor-1 {
		shouldSplit = true
	}

	if !shouldSplit {
		return nil, false
	}

	newNode := tree.createTreeNode()
	addToNew := false

	splitIdx := int(math.Floor(float64(tree.branchFactor) / 2))
	if tree.branchFactor == 2 {
		// only one of the two nodes will have room...
		if tree.compare(q, parent.splits[0]) < 0 {
			splitIdx--
		} else {
			addToNew = true
		}
	}

	if splitIdx < len(parent.splits) {
		if parent.isLeafNode() {
			newNode.splits = make([]string, len(parent.splits)-splitIdx)
			copy(newNode.splits, parent.splits[splitIdx:])

			newNode.data = make([][]interface{}, len(parent.splits)-splitIdx)
			copy(newNode.data, parent.data[splitIdx:])
			parent.data = parent.data[:splitIdx]
		} else {
			newNode.splits = make([]string, len(parent.splits)-splitIdx-1)
			copy(newNode.splits, parent.splits[splitIdx+1:])
		}

		if tree.compare(q, parent.splits[splitIdx]) >= 0 {
			addToNew = true
		}
		parent.splits = parent.splits[:splitIdx]
	} // else is branchFactor = 2 (binary tree)

	if !parent.isLeafNode() {
		childSplitIdx := splitIdx + 1
		newNode.children = make([]*bPlusTreeNode, len(parent.children)-childSplitIdx)
		copy(newNode.children, parent.children[childSplitIdx:])
		for i, child := range newNode.children {
			child.parent = newNode
			child.parentChildIdx = i
		}
		parent.children = parent.children[:childSplitIdx]
	}

	return newNode, addToNew
}

func (tree *BPlusTree) recInsert(q string, parent *bPlusTreeNode, v interface{}, child *bPlusTreeNode) {

	indent := ""
	node := parent.parent
	for node != nil {
		indent += "  "
		node = node.parent
	}

	insertIdx, performInsert := len(parent.splits), true
	for i, split := range parent.splits {

		cmpVal := tree.compare(q, split)
		if cmpVal < 0 {
			insertIdx = i
			break
		} else if cmpVal == 0 {
			if child == nil {
				// duplicate in a leaf node...
				// merge into existing data
				insertIdx = i
				performInsert = false
				break
			}
		}
	}

	nodeToInsert := parent
	var newNode *bPlusTreeNode
	addToNew := false
	if performInsert {
		newNode, addToNew = tree.splitIfNecessary(parent, q)
		if addToNew {
			nodeToInsert = newNode
			insertIdx -= len(parent.splits)
			if len(newNode.children) > 0 {
				insertIdx -= 1
			}
		}
	}

	if child != nil {

		// now I need to compare and see which one promotes the split
		if len(nodeToInsert.children) > 0 {
			nodeWithSplitValue := nodeToInsert.children[insertIdx]
			for len(nodeWithSplitValue.splits) == 0 {
				nodeWithSplitValue = nodeWithSplitValue.children[0]
			}

			cmpVal := tree.compare(q, nodeWithSplitValue.splits[0])
			if cmpVal >= 0 {
				insertIdx++
			}
		}

		if insertIdx < len(nodeToInsert.children) {
			nodeToInsert.children = append(nodeToInsert.children, nil)
			copy(nodeToInsert.children[insertIdx+1:], nodeToInsert.children[insertIdx:])
			nodeToInsert.children[insertIdx] = child
			child.parent = nodeToInsert
			child.parentChildIdx = insertIdx
			for j := insertIdx + 1; j < len(nodeToInsert.children); j++ {
				nodeToInsert.children[j].parentChildIdx = j
			}
		} else {
			nodeToInsert.children = append(nodeToInsert.children, child)
			child.parent = nodeToInsert
			child.parentChildIdx = insertIdx
		}
		insertIdx--
	} else {
		addValueToNode(nodeToInsert, v, insertIdx, !performInsert)
	}

	if performInsert && insertIdx >= 0 {
		addKeyToNode(nodeToInsert, q, insertIdx)
	}

	if newNode != nil {
		tree.addToParentNode(parent.parent, newNode)
	}
}

func addValueToNode(nodeToInsert *bPlusTreeNode, v interface{}, insertIdx int, merge bool) {
	if merge {
		if !checkIfAlreadyThere(nodeToInsert.data[insertIdx], v) {
			nodeToInsert.data[insertIdx] = append(nodeToInsert.data[insertIdx], v)
		}
	} else {
		data := make([]interface{}, 1)
		data[0] = v
		if insertIdx < len(nodeToInsert.data) {
			nodeToInsert.data = append(nodeToInsert.data, nil)
			copy(nodeToInsert.data[insertIdx+1:], nodeToInsert.data[insertIdx:])
			nodeToInsert.data[insertIdx] = data
		} else {
			nodeToInsert.data = append(nodeToInsert.data, data)
		}
	}
}

func checkIfAlreadyThere(existingValues []interface{}, value interface{}) bool {
	for _, existingValue := range existingValues {
		switch existingValue.(type) {
		case int:
			if existingValue == value {
				return true
			}
		}
	}

	return false
}

func addKeyToNode(nodeToInsert *bPlusTreeNode, q string, insertIdx int) {
	if insertIdx < len(nodeToInsert.splits) {
		nodeToInsert.splits = append(nodeToInsert.splits, " ")
		copy(nodeToInsert.splits[insertIdx+1:], nodeToInsert.splits[insertIdx:])
		nodeToInsert.splits[insertIdx] = q
		if insertIdx == 0 && nodeToInsert.parent != nil && nodeToInsert.parentChildIdx > 0 {
			// update parent split for this node in nodeToInsert.parent.splits
			nodeToInsert.parent.splits[nodeToInsert.parentChildIdx-1] = q
		}
	} else {
		nodeToInsert.splits = append(nodeToInsert.splits, q)
	}
}

func (node *bPlusTreeNode) isLeafNode() bool {
	return len(node.children) == 0
}

func (tree *BPlusTree) recFindNode(q string, node *bPlusTreeNode) *bPlusTreeNode {

	if node.isLeafNode() {
		return node
	}

	for j, split := range node.splits {
		if tree.compare(q, split) < 0 {
			return tree.recFindNode(q, node.children[j])
		}
	}
	return tree.recFindNode(q, node.children[len(node.splits)])
}
