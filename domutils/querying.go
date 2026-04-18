package domutils

import "github.com/krozhkov/go-htmlparser2/dom"

type TestType = func(elem *dom.Node) bool

/**
 * Search a node and its children for nodes passing a test function. If `node` is not an array, it will be wrapped in one.
 *
 * @category Querying
 * @param test Function to test nodes on.
 * @param node Node to search. Will be included in the result set if it matches.
 * @param recurse Also consider child nodes.
 * @param limit Maximum number of nodes to return.
 * @returns All nodes passing `test`.
 */
func Filter(
	test TestType,
	nodes []*dom.Node,
	recurse bool,
	limit int,
) []*dom.Node {
	return Find(test, nodes, recurse, limit)
}

/**
 * Search an array of nodes and their children for nodes passing a test function.
 *
 * @category Querying
 * @param test Function to test nodes on.
 * @param nodes Array of nodes to search.
 * @param recurse Also consider child nodes.
 * @param limit Maximum number of nodes to return.
 * @returns All nodes passing `test`.
 */
func Find(
	test TestType,
	nodes []*dom.Node,
	recurse bool,
	limit int,
) []*dom.Node {
	if len(nodes) == 0 {
		return nil
	}

	result := []*dom.Node{}
	/** Stack of the arrays we are looking at. */
	nodeStack := [][]*dom.Node{nodes}
	/** Stack of the indices within the arrays. */
	indexStack := []int{0}

	for {
		// First, check if the current array has any more elements to look at.
		length := len(indexStack)
		if indexStack[length-1] >= len(nodeStack[length-1]) {
			// If we have no more arrays to look at, we are done.
			if len(indexStack) == 1 {
				return result
			}

			// Otherwise, remove the current array from the stack.
			nodeStack[length-1] = nil
			nodeStack = nodeStack[:length-1]
			indexStack = indexStack[:length-1]

			// Loop back to the start to continue with the next array.
			continue
		}

		length = len(indexStack)
		elem := nodeStack[length-1][indexStack[length-1]]
		indexStack[length-1]++

		if test(elem) {
			result = append(result, elem)
			limit--
			if limit <= 0 {
				return result
			}
		}

		if recurse && dom.HasChildren(elem) && elem != nil && len(elem.Children) > 0 {
			/*
			 * Add the children to the stack. We are depth-first, so this is
			 * the next array we look at.
			 */
			indexStack = append(indexStack, 0)
			nodeStack = append(nodeStack, elem.Children)
		}
	}
}

/**
 * Finds the first element inside of an array that matches a test function. This is an alias for `Array.prototype.find`.
 *
 * @category Querying
 * @param test Function to test nodes on.
 * @param nodes Array of nodes to search.
 * @returns The first node in the array that passes `test`.
 * @deprecated Use `Array.prototype.find` directly.
 */
func FindOneChild(
	test TestType,
	nodes []*dom.Node,
) *dom.Node {
	for _, node := range nodes {
		if test(node) {
			return node
		}
	}

	return nil
}

/**
 * Finds one element in a tree that passes a test.
 *
 * @category Querying
 * @param test Function to test nodes on.
 * @param nodes Node or array of nodes to search.
 * @param recurse Also consider child nodes.
 * @returns The first node that passes `test`.
 */
func FindOne(
	test TestType,
	nodes []*dom.Node,
	recurse bool,
) *dom.Node {
	for _, node := range nodes {
		if dom.IsTag(node) && test(node) {
			return node
		}
		if recurse && dom.HasChildren(node) && len(node.Children) > 0 {
			found := FindOne(test, node.Children, true)
			if found != nil {
				return found
			}
		}
	}

	return nil
}

/**
 * Checks if a tree of nodes contains at least one node passing a test.
 *
 * @category Querying
 * @param test Function to test nodes on.
 * @param nodes Array of nodes to search.
 * @returns Whether a tree of nodes contains at least one node passing the test.
 */
func ExistsOne(
	test TestType,
	nodes []*dom.Node,
) bool {
	for _, node := range nodes {
		if (dom.IsTag(node) && test(node)) || (dom.HasChildren(node) && ExistsOne(test, node.Children)) {
			return true
		}
	}

	return false
}

/**
 * Search an array of nodes and their children for elements passing a test function.
 *
 * Same as `find`, but limited to elements and with less options, leading to reduced complexity.
 *
 * @category Querying
 * @param test Function to test nodes on.
 * @param nodes Array of nodes to search.
 * @returns All nodes passing `test`.
 */
func FindAll(
	test TestType,
	nodes []*dom.Node,
) []*dom.Node {
	result := []*dom.Node{}
	nodeStack := [][]*dom.Node{nodes}
	indexStack := []int{0}

	for {
		length := len(indexStack)
		if indexStack[length-1] >= len(nodeStack[length-1]) {
			if len(indexStack) == 1 {
				return result
			}

			// Otherwise, remove the current array from the stack.
			nodeStack[length-1] = nil
			nodeStack = nodeStack[:length-1]
			indexStack = indexStack[:length-1]

			// Loop back to the start to continue with the next array.
			continue
		}

		length = len(indexStack)
		elem := nodeStack[length-1][indexStack[length-1]]
		indexStack[length-1]++

		if dom.IsTag(elem) && test(elem) {
			result = append(result, elem)
		}

		if dom.HasChildren(elem) && elem != nil && len(elem.Children) > 0 {
			indexStack = append(indexStack, 0)
			nodeStack = append(nodeStack, elem.Children)
		}
	}
}
