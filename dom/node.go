package dom

import (
	"fmt"

	"github.com/elliotchance/orderedmap/v3"
)

/**
 * The description of an individual attribute.
 */
type Attribute = orderedmap.Element[string, string]

type Attributes = orderedmap.OrderedMap[string, string]

/**
 * This object will be used as the prototype for Nodes when creating a
 * DOM-Level-1-compliant structure.
 */
type Node struct {
	/** The type of the node. */
	Type ElementType

	/** Parent of the node */
	Parent *Node

	/** Previous sibling */
	PreviousSibling *Node

	/** Next sibling */
	NextSibling *Node

	/** The start index of the node. Requires `withStartIndices` on the handler to be `true. */
	StartIndex int

	/** The end index of the node. Requires `withEndIndices` on the handler to be `true. */
	EndIndex int

	/**
	 * [DOM spec](https://dom.spec.whatwg.org/#dom-node-nodetype)-compatible
	 * node {@link type}.
	 */
	NodeType int

	/**
	 * @param name Name of the tag, eg. `div`, `span`.
	 */
	Name string

	/**
	 * @param data The content of the data node
	 */
	Data string

	/**
	 * @param children Children of the node. Only certain node types can have children.
	 */
	Children []*Node

	/**
	 * @param attribs Object mapping attribute names to attribute values.
	 */
	Attribs *Attributes
}

func NewNode(typ ElementType, nodeType int) *Node {
	return &Node{
		Type:     typ,
		NodeType: nodeType,
	}
}

func (n *Node) CloneNode(recursive bool) (*Node, error) {
	return CloneNode(n, recursive)
}

/**
 * A node that contains some data.
 */
func NewDataNode(typ ElementType, nodeType int, data string) *Node {
	return &Node{
		Type:     typ,
		NodeType: nodeType,
		Data:     data,
	}
}

/**
 * Text within the document.
 */
func NewText(data string) *Node {
	return NewDataNode(ElementTypeText, 3, data)
}

/**
 * Comments within the document.
 */
func NewComment(data string) *Node {
	return NewDataNode(ElementTypeComment, 8, data)
}

/**
 * Processing instructions, including doc types.
 */
func NewProcessingInstruction(name string, data string) *Node {
	return &Node{
		Type:     ElementTypeDirective,
		NodeType: 1,
		Data:     data,
		Name:     name,
	}
}

/**
 * A node that can have children.
 */
func NewNodeWithChildren(typ ElementType, nodeType int, children []*Node) *Node {
	return &Node{
		Type:     typ,
		NodeType: nodeType,
		Children: children,
	}
}

// Aliases
/** First child of the node. */
func (n *Node) FirstChild() *Node {
	if len(n.Children) > 0 {
		return n.Children[0]
	}

	return nil
}

/** Last child of the node. */
func (n *Node) LastChild() *Node {
	if len(n.Children) > 0 {
		return n.Children[len(n.Children)-1]
	}

	return nil
}

/**
 * CDATA nodes.
 */
func NewCDATA(children []*Node) *Node {
	return NewNodeWithChildren(ElementTypeCDATA, 4, children)
}

/**
 * The root node of the document.
 */
func NewDocument(children []*Node) *Node {
	return NewNodeWithChildren(ElementTypeRoot, 9, children)
}

/**
 * An element within the DOM.
 */
func NewElement(name string, attribs *Attributes, children []*Node, typ ElementType) *Node {
	if typ == "" {
		typ = ElementTypeTag
		if name == "script" {
			typ = ElementTypeScript
		}
		if name == "style" {
			typ = ElementTypeStyle
		}
	}

	return &Node{
		Type:     typ,
		NodeType: 1,
		Children: children,
		Name:     name,
		Attribs:  attribs,
	}
}

/**
 * Same as {@link name}.
 * [DOM spec](https://dom.spec.whatwg.org)-compatible alias.
 */
func (n *Node) TagName() string {
	return n.Name
}

func (n *Node) Attributes() []*Attribute {
	if n.Attribs != nil {
		attributes := make([]*Attribute, 0, n.Attribs.Len())
		for name, value := range n.Attribs.AllFromFront() {
			attributes = append(attributes, &Attribute{Key: name, Value: value})
		}

		return attributes
	}

	return nil
}

/**
 * Checks if `node` is an element node.
 *
 * @param node Node to check.
 * @returns `true` if the node is an element node.
 */
func IsTag(node *Node) bool {
	return node.Type == ElementTypeTag || node.Type == ElementTypeScript || node.Type == ElementTypeStyle
}

/**
 * Checks if `node` is a CDATA node.
 *
 * @param node Node to check.
 * @returns `true` if the node is a CDATA node.
 */
func IsCDATA(node *Node) bool {
	return node.Type == ElementTypeCDATA
}

/**
 * Checks if `node` is a text node.
 *
 * @param node Node to check.
 * @returns `true` if the node is a text node.
 */
func IsText(node *Node) bool {
	return node.Type == ElementTypeText
}

/**
 * Checks if `node` is a comment node.
 *
 * @param node Node to check.
 * @returns `true` if the node is a comment node.
 */
func IsComment(node *Node) bool {
	return node.Type == ElementTypeComment
}

/**
 * Checks if `node` is a directive node.
 *
 * @param node Node to check.
 * @returns `true` if the node is a directive node.
 */
func IsDirective(node *Node) bool {
	return node.Type == ElementTypeDirective
}

/**
 * Checks if `node` is a document node.
 *
 * @param node Node to check.
 * @returns `true` if the node is a document node.
 */
func IsDocument(node *Node) bool {
	return node.Type == ElementTypeRoot
}

/**
 * Checks if `node` has children.
 *
 * @param node Node to check.
 * @returns `true` if the node has children.
 */
func HasChildren(node *Node) bool {
	return len(node.Children) > 0
}

/**
 * Clone a node, and optionally its children.
 *
 * @param recursive Clone child nodes as well.
 * @returns A clone of the node.
 */
func CloneNode(node *Node, recursive bool) (*Node, error) {
	if node == nil {
		return nil, nil
	}

	var result *Node

	if IsText(node) {
		result = NewText(node.Data)
	} else if IsComment(node) {
		result = NewComment(node.Data)
	} else if IsTag(node) {
		var children []*Node
		if recursive {
			childs, err := cloneChildren(node.Children)
			if err != nil {
				return nil, err
			}

			children = childs
		}

		clone := NewElement(node.Name, node.Attribs, children, node.Type)
		for _, child := range children {
			child.Parent = clone
		}

		result = clone
	} else if IsCDATA(node) {
		var children []*Node
		if recursive {
			childs, err := cloneChildren(node.Children)
			if err != nil {
				return nil, err
			}

			children = childs
		}

		clone := NewCDATA(children)
		for _, child := range children {
			child.Parent = clone
		}
		result = clone
	} else if IsDocument(node) {
		var children []*Node
		if recursive {
			childs, err := cloneChildren(node.Children)
			if err != nil {
				return nil, err
			}

			children = childs
		}
		clone := NewDocument(children)
		for _, child := range children {
			child.Parent = clone
		}

		result = clone
	} else if IsDirective(node) {
		instruction := NewProcessingInstruction(node.Name, node.Data)

		result = instruction
	} else {
		return nil, fmt.Errorf("Not implemented yet: %s", node.Type)
	}

	result.StartIndex = node.StartIndex
	result.EndIndex = node.EndIndex

	return result, nil
}

/**
 * Clone a list of child nodes.
 *
 * @param childs The child nodes to clone.
 * @returns A list of cloned child nodes.
 */
func cloneChildren(childs []*Node) ([]*Node, error) {
	if childs == nil {
		return nil, nil
	}

	children := make([]*Node, 0, len(childs))
	for _, child := range childs {
		clone, err := CloneNode(child, true)
		if err != nil {
			return nil, err
		}

		children = append(children, clone)
	}

	for i := 1; i < len(children); i++ {
		children[i].PreviousSibling = children[i-1]
		children[i-1].NextSibling = children[i]
	}

	return children, nil
}
