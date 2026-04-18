package serializer

import (
	"strings"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/entities"
)

type DomSerializerOptions struct {
	/**
	 * Print an empty attribute's value.
	 *
	 * @default xmlMode
	 * @example With <code>emptyAttrs: false</code>: <code>&lt;input checked&gt;</code>
	 * @example With <code>emptyAttrs: true</code>: <code>&lt;input checked=""&gt;</code>
	 */
	EmptyAttrs *bool
	/**
	 * Print self-closing tags for tags without contents. If `xmlMode` is set, this will apply to all tags.
	 * Otherwise, only tags that are defined as self-closing in the HTML specification will be printed as such.
	 *
	 * @default xmlMode
	 * @example With <code>selfClosingTags: false</code>: <code>&lt;foo&gt;&lt;/foo&gt;&lt;br&gt;&lt;/br&gt;</code>
	 * @example With <code>xmlMode: true</code> and <code>selfClosingTags: true</code>: <code>&lt;foo/&gt;&lt;br/&gt;</code>
	 * @example With <code>xmlMode: false</code> and <code>selfClosingTags: true</code>: <code>&lt;foo&gt;&lt;/foo&gt;&lt;br /&gt;</code>
	 */
	SelfClosingTags *bool
	/**
	 * Treat the input as an XML document; enables the `emptyAttrs` and `selfClosingTags` options.
	 *
	 * If the value is `"foreign"`, it will try to correct mixed-case attribute names.
	 *
	 * @default false
	 */
	XmlMode *string // "false" | "true" | "foreign"
	/**
	 * Encode characters that are either reserved in HTML or XML.
	 *
	 * If `xmlMode` is `true` or the value not `'utf8'`, characters outside of the utf8 range will be encoded as well.
	 *
	 * @default `decodeEntities`
	 */
	EncodeEntities *string // "false" | "true" | "utf8"
	/**
	 * Option inherited from parsing; will be used as the default value for `encodeEntities`.
	 *
	 * @default true
	 */
	DecodeEntities *bool
}

func ptr[T any](v T) *T {
	return &v
}

func isUnencodedElements(name string) bool {
	switch name {
	case "style",
		"script",
		"xmp",
		"iframe",
		"noembed",
		"noframes",
		"plaintext",
		"noscript":
		return true
	default:
		return false
	}
}

func replaceQuotes(value string) string {
	return strings.ReplaceAll(value, "\"", "&quot;")
}

/**
 * Format attributes
 */
func formatAttributes(
	sb *strings.Builder,
	attributes *orderedmap.OrderedMap[string, string],
	opts DomSerializerOptions,
) {
	if attributes == nil {
		return
	}

	var encode func(string) string
	if (opts.EncodeEntities != nil && *opts.EncodeEntities == "false") || (opts.DecodeEntities != nil && !*opts.DecodeEntities) {
		encode = replaceQuotes
	} else if (opts.XmlMode != nil && (*opts.XmlMode == "true" || *opts.XmlMode == "foreign")) || opts.EncodeEntities == nil || *opts.EncodeEntities != "utf8" {
		encode = entities.EncodeXML
	} else {
		encode = entities.EscapeAttribute
	}

	for key, value := range attributes.AllFromFront() {
		if opts.XmlMode != nil && *opts.XmlMode == "foreign" {
			/* Fix up mixed-case attribute names */
			if fixedKey, ok := attributeNames[key]; ok {
				key = fixedKey
			}
		}

		sb.WriteString(" ")

		if (opts.EmptyAttrs == nil || !*opts.EmptyAttrs) && (opts.XmlMode == nil || *opts.XmlMode == "false") && value == "" {
			sb.WriteString(key)
		} else {
			sb.WriteString(key)
			sb.WriteString("=\"")
			sb.WriteString(encode(value))
			sb.WriteString("\"")
		}
	}
}

/**
 * Self-enclosing tags
 */
func isSingleTag(name string) bool {
	switch name {
	case "area",
		"base",
		"basefont",
		"br",
		"col",
		"command",
		"embed",
		"frame",
		"hr",
		"img",
		"input",
		"isindex",
		"keygen",
		"link",
		"meta",
		"param",
		"source",
		"track",
		"wbr":
		return true
	default:
		return false
	}
}

/**
 * Renders a DOM node or an array of DOM nodes to a string.
 *
 * Can be thought of as the equivalent of the `outerHTML` of the passed node(s).
 *
 * @param node Node to be rendered.
 * @param options Changes serialization behavior
 */
func Render(
	nodes []*dom.Node,
	options *DomSerializerOptions,
) string {
	var sb = new(strings.Builder)

	if options == nil {
		options = &DomSerializerOptions{}
	}

	render(sb, nodes, *options)

	return sb.String()
}

func render(
	sb *strings.Builder,
	nodes []*dom.Node,
	options DomSerializerOptions,
) {
	for i := 0; i < len(nodes); i++ {
		if nodes[i] != nil {
			renderNode(sb, nodes[i], options)
		}
	}
}

func renderNode(sb *strings.Builder, node *dom.Node, options DomSerializerOptions) {
	switch node.Type {
	case dom.ElementTypeRoot:
		render(sb, node.Children, options)
	case dom.ElementTypeDoctype:
		renderDirective(sb, node)
	case dom.ElementTypeDirective:
		renderDirective(sb, node)
	case dom.ElementTypeComment:
		renderComment(sb, node)
	case dom.ElementTypeCDATA:
		renderCdata(sb, node)
	case dom.ElementTypeScript:
		renderTag(sb, node, options)
	case dom.ElementTypeStyle:
		renderTag(sb, node, options)
	case dom.ElementTypeTag:
		renderTag(sb, node, options)
	case dom.ElementTypeText:
		renderText(sb, node, options)
	}
}

func isForeignModeIntegrationPoints(name string) bool {
	switch name {
	case "mi",
		"mo",
		"mn",
		"ms",
		"mtext",
		"annotation-xml",
		"foreignObject",
		"desc",
		"title":
		return true
	default:
		return false
	}
}

func isForeignElements(name string) bool {
	switch name {
	case "svg", "math":
		return true
	default:
		return false
	}
}

func renderTag(sb *strings.Builder, elem *dom.Node, opts DomSerializerOptions) {
	// Handle SVG / MathML in HTML
	if opts.XmlMode != nil && *opts.XmlMode == "foreign" {
		/* Fix up mixed-case element names */
		if fixedName, ok := elementNames[elem.Name]; ok {
			elem.Name = fixedName
		}
		/* Exit foreign mode at integration points */
		if elem.Parent != nil && isForeignModeIntegrationPoints(elem.Parent.Name) {
			opts.XmlMode = ptr("false")
		}
	}
	if (opts.XmlMode == nil || *opts.XmlMode == "false") && isForeignElements(elem.Name) {
		opts.XmlMode = ptr("foreign")
	}

	sb.WriteString("<")
	sb.WriteString(elem.Name)
	formatAttributes(sb, elem.Attribs, opts)

	var selfClosing bool
	if opts.XmlMode != nil && (*opts.XmlMode == "true" || *opts.XmlMode == "foreign") {
		// In XML mode or foreign mode, and user hasn't explicitly turned off self-closing tags
		if opts.SelfClosingTags == nil || *opts.SelfClosingTags {
			selfClosing = true
		}
	} else {
		// User explicitly asked for self-closing tags, even in HTML mode
		if opts.SelfClosingTags != nil && *opts.SelfClosingTags && isSingleTag(elem.Name) {
			selfClosing = true
		}
	}

	if len(elem.Children) == 0 && selfClosing {
		if opts.XmlMode == nil || *opts.XmlMode == "false" {
			sb.WriteString(" ")
		}
		sb.WriteString("/>")
	} else {
		sb.WriteString(">")
		if len(elem.Children) > 0 {
			render(sb, elem.Children, opts)
		}

		if (opts.XmlMode != nil && (*opts.XmlMode == "true" || *opts.XmlMode == "foreign")) || !isSingleTag(elem.Name) {
			sb.WriteString("</")
			sb.WriteString(elem.Name)
			sb.WriteString(">")
		}
	}
}

func renderDirective(sb *strings.Builder, elem *dom.Node) {
	sb.WriteString("<")
	sb.WriteString(elem.Data)
	sb.WriteString(">")
}

func renderText(sb *strings.Builder, elem *dom.Node, opts DomSerializerOptions) {
	// If entities weren't decoded, no need to encode them back
	var encodeEntities bool
	if opts.EncodeEntities != nil {
		encodeEntities = *opts.EncodeEntities == "true" || *opts.EncodeEntities == "utf8"
	} else {
		encodeEntities = opts.DecodeEntities == nil || *opts.DecodeEntities
	}
	var isUnencoded = (opts.XmlMode == nil || *opts.XmlMode == "false") && elem.Parent != nil && isUnencodedElements(elem.Parent.Name)
	if encodeEntities && !isUnencoded {
		if (opts.XmlMode != nil && (*opts.XmlMode == "true" || *opts.XmlMode == "foreign")) || opts.EncodeEntities == nil || *opts.EncodeEntities != "utf8" {
			sb.WriteString(entities.EncodeXML(elem.Data))
		} else {
			sb.WriteString(entities.EscapeText(elem.Data))
		}
	} else {
		sb.WriteString(elem.Data)
	}
}

func renderCdata(sb *strings.Builder, elem *dom.Node) {
	sb.WriteString("<![CDATA[")
	if len(elem.Children) > 0 {
		sb.WriteString(elem.Children[0].Data)
	}
	sb.WriteString("]]>")
}

func renderComment(sb *strings.Builder, elem *dom.Node) {
	sb.WriteString("<!--")
	sb.WriteString(elem.Data)
	sb.WriteString("-->")
}
