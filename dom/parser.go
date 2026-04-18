package dom

import (
	"unsafe"

	"github.com/krozhkov/go-htmlparser2/parser"
)

/**
 * Parses the data, returns the resulting document.
 *
 * @param data The data that should be parsed.
 * @param options Optional options for the parser and DOM handler.
 */
func ParseDocument(data string, options *parser.ParserOptions) *Node {
	var domHandlerOptions *DomHandlerOptions
	if options != nil {
		domHandlerOptions = &DomHandlerOptions{XmlMode: options.XmlMode}
	}
	handler := NewDomHandler(nil, domHandlerOptions, nil)
	parser := parser.NewParser(handler, options)
	parser.End(unsafe.Slice(unsafe.StringData(data), len(data)))
	return handler.Root
}

/**
 * Parses data, returns an array of the root nodes.
 *
 * Note that the root nodes still have a `Document` node as their parent.
 * Use `parseDocument` to get the `Document` node instead.
 *
 * @param data The data that should be parsed.
 * @param options Optional options for the parser and DOM handler.
 * @deprecated Use `parseDocument` instead.
 */
func ParseDOM(data string, options *parser.ParserOptions) []*Node {
	node := ParseDocument(data, options)
	return node.Children
}
