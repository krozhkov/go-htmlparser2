package domutils

import (
	"slices"

	"github.com/krozhkov/go-htmlparser2/dom"
)

/**
 * Remove an element from the dom
 *
 * @category Manipulation
 * @param elem The element to be removed
 */
func RemoveElement(elem *dom.Node) {
	if elem.PreviousSibling != nil {
		elem.PreviousSibling.NextSibling = elem.NextSibling
	}
	if elem.NextSibling != nil {
		elem.NextSibling.PreviousSibling = elem.PreviousSibling
	}

	if elem.Parent != nil {
		childsIndex := lastIndex(elem.Parent.Children, elem)
		if childsIndex >= 0 {
			elem.Parent.Children = slices.Delete(elem.Parent.Children, childsIndex, childsIndex+1)
		}
	}
	elem.NextSibling = nil
	elem.PreviousSibling = nil
	elem.Parent = nil
}

/**
 * Replace an element in the dom
 *
 * @category Manipulation
 * @param elem The element to be replaced
 * @param replacement The element to be added
 */
func ReplaceElement(elem *dom.Node, replacement *dom.Node) {
	prev := elem.PreviousSibling
	replacement.PreviousSibling = prev
	if prev != nil {
		prev.NextSibling = replacement
		elem.PreviousSibling = nil
	}

	next := elem.NextSibling
	replacement.NextSibling = next
	if next != nil {
		next.PreviousSibling = replacement
		elem.NextSibling = nil
	}

	parent := elem.Parent
	replacement.Parent = parent
	if parent != nil {
		childs := parent.Children
		childsIndex := lastIndex(childs, elem)
		childs[childsIndex] = replacement
		elem.Parent = nil
	}
}

/**
 * Append a child to an element.
 *
 * @category Manipulation
 * @param parent The element to append to.
 * @param child The element to be added as a child.
 */
func AppendChild(parent *dom.Node, child *dom.Node) {
	RemoveElement(child)

	child.Parent = parent
	parent.Children = append(parent.Children, child)

	if len(parent.Children) > 1 {
		sibling := parent.Children[len(parent.Children)-2]
		sibling.NextSibling = child
		child.PreviousSibling = sibling
	}
}

/**
 * Append an element after another.
 *
 * @category Manipulation
 * @param elem The element to append after.
 * @param next The element be added.
 */
func Append(elem *dom.Node, next *dom.Node) {
	RemoveElement(next)

	parent := elem.Parent
	currNext := elem.NextSibling

	next.NextSibling = currNext
	next.PreviousSibling = elem
	elem.NextSibling = next
	next.Parent = parent

	if currNext != nil {
		currNext.PreviousSibling = next
		if parent != nil {
			index := lastIndex(parent.Children, currNext)
			parent.Children = slices.Insert(parent.Children, index, next)
		}
	} else if parent != nil {
		parent.Children = append(parent.Children, next)
	}
}

/**
 * Prepend a child to an element.
 *
 * @category Manipulation
 * @param parent The element to prepend before.
 * @param child The element to be added as a child.
 */
func PrependChild(parent *dom.Node, child *dom.Node) {
	RemoveElement(child)

	child.Parent = parent
	parent.Children = slices.Insert(parent.Children, 0, child)

	if len(parent.Children) > 1 {
		sibling := parent.Children[1]
		sibling.PreviousSibling = child
		child.NextSibling = sibling
	}
}

/**
 * Prepend an element before another.
 *
 * @category Manipulation
 * @param elem The element to prepend before.
 * @param prev The element be added.
 */
func Prepend(elem *dom.Node, prev *dom.Node) {
	RemoveElement(prev)

	parent := elem.Parent
	if parent != nil {
		index := lastIndex(parent.Children, elem)
		parent.Children = slices.Insert(parent.Children, index, prev)
	}

	if elem.PreviousSibling != nil {
		elem.PreviousSibling.NextSibling = prev
	}

	prev.Parent = parent
	prev.PreviousSibling = elem.PreviousSibling
	prev.NextSibling = elem
	elem.PreviousSibling = prev
}
