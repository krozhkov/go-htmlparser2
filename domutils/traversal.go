package domutils

import (
	"slices"

	"github.com/krozhkov/go-htmlparser2/dom"
)

/**
 * Get a node's children.
 *
 * @category Traversal
 * @param elem Node to get the children of.
 * @returns `elem`'s children, or an empty array.
 */
func GetChildren(elem *dom.Node) []*dom.Node {
	if elem != nil && dom.HasChildren(elem) {
		return elem.Children
	}

	return nil
}

/**
 * Get a node's parent.
 *
 * @category Traversal
 * @param elem Node to get the parent of.
 * @returns `elem`'s parent node, or `null` if `elem` is a root node.
 */
func GetParent(elem *dom.Node) *dom.Node {
	if elem != nil {
		return elem.Parent
	}
	return nil
}

/**
 * Gets an elements siblings, including the element itself.
 *
 * Attempts to get the children through the element's parent first. If we don't
 * have a parent (the element is a root node), we walk the element's `prev` &
 * `next` to get all remaining nodes.
 *
 * @category Traversal
 * @param elem Element to get the siblings of.
 * @returns `elem`'s siblings, including `elem`.
 */
func GetSiblings(elem *dom.Node) []*dom.Node {
	if elem == nil {
		return nil
	}

	parent := GetParent(elem)
	if parent != nil {
		return GetChildren(parent)
	}

	siblings := []*dom.Node{elem}
	prev := elem.PreviousSibling
	for prev != nil {
		siblings = append(siblings, prev)
		prev = prev.PreviousSibling
	}
	slices.Reverse(siblings)

	next := elem.NextSibling
	for next != nil {
		siblings = append(siblings, next)
		next = next.NextSibling
	}
	return siblings
}

/**
 * Gets an attribute from an element.
 *
 * @category Traversal
 * @param elem Element to check.
 * @param name Attribute name to retrieve.
 * @returns The element's attribute value, or `undefined`.
 */
func GetAttributeValue(elem *dom.Node, name string) string {
	if elem != nil && elem.Attribs != nil {
		return elem.Attribs.GetOrDefault(name, "")
	}
	return ""
}

/**
 * Checks whether an element has an attribute.
 *
 * @category Traversal
 * @param elem Element to check.
 * @param name Attribute name to look for.
 * @returns Returns whether `elem` has the attribute `name`.
 */
func HasAttrib(elem *dom.Node, name string) bool {
	return elem != nil && elem.Attribs != nil && elem.Attribs.Has(name)
}

/**
 * Get the tag name of an element.
 *
 * @category Traversal
 * @param elem The element to get the name for.
 * @returns The tag name of `elem`.
 */
func GetName(elem *dom.Node) string {
	if elem != nil {
		return elem.Name
	}
	return ""
}

/**
 * Returns the next element sibling of a node.
 *
 * @category Traversal
 * @param elem The element to get the next sibling of.
 * @returns `elem`'s next sibling that is a tag, or `null` if there is no next
 * sibling.
 */
func NextElementSibling(elem *dom.Node) *dom.Node {
	if elem == nil {
		return nil
	}

	next := elem.NextSibling
	for next != nil && !dom.IsTag(next) {
		next = next.NextSibling
	}
	return next
}

/**
 * Returns the previous element sibling of a node.
 *
 * @category Traversal
 * @param elem The element to get the previous sibling of.
 * @returns `elem`'s previous sibling that is a tag, or `null` if there is no
 * previous sibling.
 */
func PrevElementSibling(elem *dom.Node) *dom.Node {
	if elem == nil {
		return nil
	}

	prev := elem.PreviousSibling
	for prev != nil && !dom.IsTag(prev) {
		prev = prev.PreviousSibling
	}
	return prev
}
