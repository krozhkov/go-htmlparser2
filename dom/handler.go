package dom

import (
	"github.com/elliotchance/orderedmap/v3"
	"github.com/krozhkov/go-htmlparser2/parser"
)

type DomHandlerOptions struct {
	/**
	 * Add a `startIndex` property to nodes.
	 * When the parser is used in a non-streaming fashion, `startIndex` is an integer
	 * indicating the position of the start of the node in the document.
	 *
	 * @default false
	 */
	WithStartIndices bool

	/**
	 * Add an `endIndex` property to nodes.
	 * When the parser is used in a non-streaming fashion, `endIndex` is an integer
	 * indicating the position of the end of the node in the document.
	 *
	 * @default false
	 */
	WithEndIndices bool

	/**
	 * Treat the markup as XML.
	 *
	 * @default false
	 */
	XmlMode bool
}

type Callback func(err error, dom []*Node)
type ElementCallback func(element *Node)

type DomHandler struct {
	/** The elements of the DOM */
	Dom []*Node
	/** The root element for the DOM */
	Root *Node

	/** Called once parsing has completed. */
	callback Callback
	/** Settings for the handler. */
	options *DomHandlerOptions
	/** Callback whenever a tag is closed. */
	elementCB ElementCallback
	/** Indicated whether parsing has been completed. */
	done bool
	/** Stack of open tags. */
	tagStack []*Node
	/** A data node that is still being written to. */
	lastNode *Node
	/** Reference to the parser instance. Used for location information. */
	parser *parser.Parser
}

func NewDomHandler(callback Callback, options *DomHandlerOptions, elementCB ElementCallback) *DomHandler {
	dom := make([]*Node, 0, 10)
	root := NewDocument(dom)
	return &DomHandler{
		Dom:       dom,
		Root:      root,
		callback:  callback,
		options:   options,
		elementCB: elementCB,
		done:      false,
		tagStack:  []*Node{root},
	}
}

func (h *DomHandler) OnParserInit(parser *parser.Parser) {
	h.parser = parser
}

func (h *DomHandler) OnReset() {
	h.Dom = make([]*Node, 0, 10)
	h.Root = NewDocument(h.Dom)
	h.done = false
	h.tagStack = []*Node{h.Root}
	h.lastNode = nil
	h.parser = nil
}

func (h *DomHandler) OnEnd() {
	if h.done {
		return
	}

	h.done = true
	h.parser = nil
	h.handleCallback(nil)
}

func (h *DomHandler) OnError(e error) {
	h.handleCallback(e)
}

func (h *DomHandler) OnCloseTag(name string, isImplied bool) {
	if h.parser == nil {
		return
	}

	h.lastNode = nil

	var elem *Node
	if len(h.tagStack) > 0 {
		lastIndex := len(h.tagStack) - 1
		elem = h.tagStack[lastIndex]
		h.tagStack[lastIndex] = nil
		h.tagStack = h.tagStack[:lastIndex]
	}

	if h.options != nil && h.options.WithEndIndices && elem != nil {
		elem.EndIndex = h.parser.EndIndex
	}

	if h.elementCB != nil && elem != nil {
		h.elementCB(elem)
	}
}

func (h *DomHandler) OnOpenTag(name string, attrs []*parser.Attribute, isImplied bool) {
	var typ ElementType = ""
	if h.options != nil && h.options.XmlMode {
		typ = ElementTypeTag
	}

	attributes := orderedmap.NewOrderedMapWithCapacity[string, string](len(attrs))
	for _, el := range attrs {
		attributes.Set(el.Name, el.Value)
	}
	element := NewElement(name, attributes, nil, typ)
	h.addNode(element)
	h.tagStack = append(h.tagStack, element)
}

func (h *DomHandler) OnText(data string) {
	lastNode := h.lastNode

	if lastNode != nil && lastNode.Type == ElementTypeText {
		lastNode.Data += data
		if h.options != nil && h.options.WithEndIndices {
			lastNode.EndIndex = h.parser.EndIndex
		}
	} else {
		node := NewText(data)
		h.addNode(node)
		h.lastNode = node
	}
}

func (h *DomHandler) OnComment(data string) {
	if h.lastNode != nil && h.lastNode.Type == ElementTypeComment {
		h.lastNode.Data += data
		return
	}

	node := NewComment(data)
	h.addNode(node)
	h.lastNode = node
}

func (h *DomHandler) OnCommentEnd() {
	h.lastNode = nil
}

func (h *DomHandler) OnCDataStart() {
	text := NewText("")
	node := NewCDATA([]*Node{text})

	h.addNode(node)

	text.Parent = node
	h.lastNode = text
}

func (h *DomHandler) OnCDataEnd() {
	h.lastNode = nil
}

func (h *DomHandler) OnProcessingInstruction(name string, data string) {
	node := NewProcessingInstruction(name, data)
	h.addNode(node)
}

func (h *DomHandler) OnOpenTagName(name string) {
}

func (h *DomHandler) OnAttribute(name string, value string, quote parser.QuoteType) {
}

func (h *DomHandler) handleCallback(err error) {
	if h.callback != nil {
		h.callback(err, h.Dom)
	} else if err != nil {
		panic(err)
	}
}

func (h *DomHandler) addNode(node *Node) {
	var parent, previousSibling *Node
	if len(h.tagStack) > 0 {
		parent = h.tagStack[len(h.tagStack)-1]
	}

	if h.parser == nil || parent == nil || node == nil {
		return
	}

	if HasChildren(parent) {
		children := parent.Children
		if len(children) > 0 {
			previousSibling = children[len(children)-1]
		}
	}

	if h.options != nil && h.options.WithStartIndices {
		node.StartIndex = h.parser.StartIndex
	}

	if h.options != nil && h.options.WithEndIndices {
		node.EndIndex = h.parser.EndIndex
	}

	parent.Children = append(parent.Children, node)

	if previousSibling != nil {
		node.PreviousSibling = previousSibling
		previousSibling.NextSibling = node
	}

	node.Parent = parent
	h.lastNode = nil
}
