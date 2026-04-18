package domutils

import "github.com/krozhkov/go-htmlparser2/dom"

/**
 * Returns the node with the supplied ID.
 *
 * @category Legacy Query Functions
 * @param id The unique ID attribute value to look for.
 * @param nodes Nodes to search through.
 * @param recurse Also consider child nodes.
 * @returns The node with the supplied ID.
 */
func GetElementById(
	id string,
	nodes []*dom.Node,
	recurse bool,
) *dom.Node {
	return FindOne(
		func(elem *dom.Node) bool {
			return elem != nil && dom.IsTag(elem) && elem.Attribs.GetOrDefault("id", "") == id
		},
		nodes,
		recurse,
	)
}

func GetElementByIdFn(
	test func(id string) bool,
	nodes []*dom.Node,
	recurse bool,
) *dom.Node {
	return FindOne(
		func(elem *dom.Node) bool {
			return elem != nil && dom.IsTag(elem) && test(elem.Attribs.GetOrDefault("id", ""))
		},
		nodes,
		recurse,
	)
}

/**
 * Returns all nodes with the supplied `tagName`.
 *
 * @category Legacy Query Functions
 * @param tagName Tag name to search for.
 * @param nodes Nodes to search through.
 * @param recurse Also consider child nodes.
 * @param limit Maximum number of nodes to return.
 * @returns All nodes with the supplied `tagName`.
 */
func GetElementsByTagName(
	tagName string,
	nodes []*dom.Node,
	recurse bool,
	limit int,
) []*dom.Node {
	return Filter(
		func(elem *dom.Node) bool {
			if elem != nil && elem.Name == "*" {
				return true
			}
			return elem != nil && dom.IsTag(elem) && elem.Name == tagName
		},
		nodes,
		recurse,
		limit,
	)
}

func GetElementsByTagNameFn(
	test func(tagName string) bool,
	nodes []*dom.Node,
	recurse bool,
	limit int,
) []*dom.Node {
	return Filter(
		func(elem *dom.Node) bool {
			return elem != nil && dom.IsTag(elem) && test(elem.Name)
		},
		nodes,
		recurse,
		limit,
	)
}

/**
 * Returns all nodes with the supplied `className`.
 *
 * @category Legacy Query Functions
 * @param className Class name to search for.
 * @param nodes Nodes to search through.
 * @param recurse Also consider child nodes.
 * @param limit Maximum number of nodes to return.
 * @returns All nodes with the supplied `className`.
 */
func GetElementsByClassName(
	className string,
	nodes []*dom.Node,
	recurse bool,
	limit int,
) []*dom.Node {
	return Filter(
		func(elem *dom.Node) bool {
			return elem != nil && dom.IsTag(elem) && elem.Attribs.GetOrDefault("class", "") == className
		},
		nodes,
		recurse,
		limit,
	)
}

func GetElementsByClassNameFn(
	test func(className string) bool,
	nodes []*dom.Node,
	recurse bool,
	limit int,
) []*dom.Node {
	return Filter(
		func(elem *dom.Node) bool {
			return elem != nil && dom.IsTag(elem) && test(elem.Attribs.GetOrDefault("class", ""))
		},
		nodes,
		recurse,
		limit,
	)
}

/**
 * Returns all nodes with the supplied `type`.
 *
 * @category Legacy Query Functions
 * @param type Element type to look for.
 * @param nodes Nodes to search through.
 * @param recurse Also consider child nodes.
 * @param limit Maximum number of nodes to return.
 * @returns All nodes with the supplied `type`.
 */
func GetElementsByTagType(
	typ dom.ElementType,
	nodes []*dom.Node,
	recurse bool,
	limit int,
) []*dom.Node {
	return Filter(
		func(elem *dom.Node) bool {
			return elem != nil && elem.Type == typ
		},
		nodes,
		recurse,
		limit,
	)
}

func GetElementsByTagTypeFn(
	test func(typ dom.ElementType) bool,
	nodes []*dom.Node,
	recurse bool,
	limit int,
) []*dom.Node {
	return Filter(
		func(elem *dom.Node) bool {
			return elem != nil && test(elem.Type)
		},
		nodes,
		recurse,
		limit,
	)
}
