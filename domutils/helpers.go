package domutils

import (
	"slices"

	"github.com/krozhkov/go-htmlparser2/dom"
)

/**
 * Given an array of nodes, remove any member that is contained by another
 * member.
 *
 * @category Helpers
 * @param nodes Nodes to filter.
 * @returns Remaining nodes that aren't contained by other nodes.
 */
func RemoveSubsets(nodes []*dom.Node) []*dom.Node {
	/*
	 * Check if each node (or one of its ancestors) is already contained in the
	 * array.
	 */
	for idx := len(nodes) - 1; idx >= 0; idx-- {
		node := nodes[idx]

		/*
		 * Remove the node if it is not unique.
		 * We are going through the array from the end, so we only
		 * have to check nodes that preceed the node under consideration in the array.
		 */
		if idx > 0 && lastIndex(nodes[:idx], node) >= 0 {
			nodes = slices.Delete(nodes, idx, idx+1)
			continue
		}

		for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
			if slices.Index(nodes, ancestor) >= 0 {
				nodes = slices.Delete(nodes, idx, idx+1)
				break
			}
		}
	}

	return nodes
}
