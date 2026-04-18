package parser

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

type Attribute struct {
	Name  string
	Value string
}

func isFormTags(name string) bool {
	switch name {
	case "input",
		"option",
		"optgroup",
		"select",
		"button",
		"datalist",
		"textarea":
		return true
	default:
		return false
	}
}
func isPTag(name string) bool {
	return name == "p"
}
func isTableSectionTag(name string) bool {
	return name == "thead" || name == "tbody"
}
func isDdtTag(name string) bool {
	return name == "dd" || name == "dt"
}
func isRtpTag(name string) bool {
	return name == "rt" || name == "rp"
}

var openImpliesClose = map[string]func(string) bool{
	"tr":         func(name string) bool { return name == "tr" || name == "th" || name == "td" },
	"th":         func(name string) bool { return name == "th" },
	"td":         func(name string) bool { return name == "thead" || name == "th" || name == "td" },
	"body":       func(name string) bool { return name == "head" || name == "link" || name == "script" },
	"li":         func(name string) bool { return name == "li" },
	"p":          isPTag,
	"h1":         isPTag,
	"h2":         isPTag,
	"h3":         isPTag,
	"h4":         isPTag,
	"h5":         isPTag,
	"h6":         isPTag,
	"select":     isFormTags,
	"input":      isFormTags,
	"output":     isFormTags,
	"button":     isFormTags,
	"datalist":   isFormTags,
	"textarea":   isFormTags,
	"option":     func(name string) bool { return name == "option" },
	"optgroup":   func(name string) bool { return name == "optgroup" || name == "option" },
	"dd":         isDdtTag,
	"dt":         isDdtTag,
	"address":    isPTag,
	"article":    isPTag,
	"aside":      isPTag,
	"blockquote": isPTag,
	"details":    isPTag,
	"div":        isPTag,
	"dl":         isPTag,
	"fieldset":   isPTag,
	"figcaption": isPTag,
	"figure":     isPTag,
	"footer":     isPTag,
	"form":       isPTag,
	"header":     isPTag,
	"hr":         isPTag,
	"main":       isPTag,
	"nav":        isPTag,
	"ol":         isPTag,
	"pre":        isPTag,
	"section":    isPTag,
	"table":      isPTag,
	"ul":         isPTag,
	"rt":         isRtpTag,
	"rp":         isRtpTag,
	"tbody":      isTableSectionTag,
	"tfoot":      isTableSectionTag,
}

func isForeignContextElements(name string) bool {
	switch name {
	case "math", "svg":
		return true
	default:
		return false
	}
}

func isHtmlIntegrationElements(name string) bool {
	switch name {
	case "mi",
		"mo",
		"mn",
		"ms",
		"mtext",
		"annotation-xml",
		"foreignobject",
		"desc",
		"title":
		return true
	default:
		return false
	}
}

type ParserOptions struct {
	/**
	 * Indicates whether special tags (`<script>`, `<style>`, and `<title>`) should get special treatment
	 * and if "empty" tags (eg. `<br>`) can have children.  If `false`, the content of special tags
	 * will be text only. For feeds and other XML content (documents that don't consist of HTML),
	 * set this to `true`.
	 *
	 * @default false
	 */
	XmlMode bool

	/**
	 * Decode entities within the document.
	 *
	 * @default true
	 */
	DecodeEntities bool

	/**
	 * If set to true, all tags will be lowercased.
	 *
	 * @default !xmlMode
	 */
	LowerCaseTags bool

	/**
	 * If set to `true`, all attribute names will be lowercased. This has noticeable impact on speed.
	 *
	 * @default !xmlMode
	 */
	LowerCaseAttributeNames bool

	/**
	 * If set to true, CDATA sections will be recognized as text even if the xmlMode option is not enabled.
	 * NOTE: If xmlMode is set to `true` then CDATA sections will always be recognized as text.
	 *
	 * @default xmlMode
	 */
	RecognizeCDATA bool

	/**
	 * If set to `true`, self-closing tags will trigger the onclosetag event even if xmlMode is not set to `true`.
	 * NOTE: If xmlMode is set to `true` then self-closing tags will always be recognized.
	 *
	 * @default xmlMode
	 */
	RecognizeSelfClosing bool

	/**
	 * Allows the default Tokenizer to be overwritten.
	 */
	Tokenizer *Tokenizer
}

type Handler interface {
	OnParserInit(parser *Parser)

	/**
	 * Resets the handler back to starting state
	 */
	OnReset()

	/**
	 * Signals the handler that parsing is done
	 */
	OnEnd()
	OnError(e error)
	OnCloseTag(name string, isImplied bool)
	OnOpenTagName(name string)
	/**
	 *
	 * @param name Name of the attribute
	 * @param value Value of the attribute.
	 * @param quote Quotes used around the attribute.
	 */
	OnAttribute(
		name string,
		value string,
		quote QuoteType,
	)
	OnOpenTag(
		name string,
		attribs []*Attribute,
		isImplied bool,
	)
	OnText(data string)
	OnComment(data string)
	OnCDataStart()
	OnCDataEnd()
	OnCommentEnd()
	OnProcessingInstruction(name string, data string)
}

type Parser struct {
	/** The start index of the last event. */
	StartIndex int
	/** The end index of the last event. */
	EndIndex int
	/**
	 * Store the start index of the current open tag,
	 * so we can update the start index for attributes.
	 */
	openTagStart int

	tagname     string
	attribname  string
	attribvalue []byte
	attribs     []*Attribute
	stack       []string
	/** Determines whether self-closing tags are recognized. */
	foreignContext          []bool
	cbs                     Handler
	lowerCaseTagNames       bool
	lowerCaseAttributeNames bool
	recognizeSelfClosing    bool
	recognizeCDATA          bool
	/** We are parsing HTML. Inverse of the `xmlMode` option. */
	htmlMode  bool
	tokenizer *Tokenizer

	buffers      [][]byte
	bufferOffset int
	/** The index of the last written buffer. Used when resuming after a `pause()`. */
	writeIndex int
	/** Indicates whether the parser has finished running / `.end` has been called. */
	ended bool
}

func NewParser(cbs Handler, options *ParserOptions) *Parser {
	var xmlMode bool
	var decodeEntities bool
	var lowerCaseTags bool
	var lowerCaseAttributeNames bool
	var recognizeCDATA bool
	var recognizeSelfClosing bool
	var tokenizer *Tokenizer
	if options != nil {
		xmlMode = options.XmlMode
		decodeEntities = options.DecodeEntities
		lowerCaseTags = options.LowerCaseTags
		lowerCaseAttributeNames = options.LowerCaseAttributeNames
		recognizeCDATA = options.RecognizeCDATA
		recognizeSelfClosing = options.RecognizeSelfClosing
		tokenizer = options.Tokenizer
	}

	p := &Parser{
		StartIndex:              0,
		EndIndex:                0,
		openTagStart:            0,
		tagname:                 "",
		attribname:              "",
		attribvalue:             nil,
		attribs:                 nil,
		stack:                   make([]string, 0, 10),
		foreignContext:          []bool{xmlMode},
		cbs:                     cbs,
		lowerCaseTagNames:       lowerCaseTags,
		lowerCaseAttributeNames: lowerCaseAttributeNames,
		recognizeSelfClosing:    recognizeSelfClosing,
		recognizeCDATA:          recognizeCDATA,
		htmlMode:                !xmlMode,
		tokenizer:               nil,
		buffers:                 make([][]byte, 0, 10),
		bufferOffset:            0,
		writeIndex:              0,
		ended:                   false,
	}

	if tokenizer != nil {
		p.tokenizer = tokenizer
	} else {
		p.tokenizer = NewTokenizer(
			TokenizerOptions{XmlMode: xmlMode, DecodeEntities: decodeEntities},
			p,
		)
	}

	p.cbs.OnParserInit(p)

	return p
}

// Tokenizer event handlers

/** @internal */
func (p *Parser) OnText(start, endIndex int) {
	data := p.getSlice(start, endIndex)
	p.EndIndex = endIndex - 1
	p.cbs.OnText(string(data))
	p.StartIndex = endIndex
}

/** @internal */
func (p *Parser) OnTextEntity(cp rune, endIndex int) {
	p.EndIndex = endIndex - 1
	p.cbs.OnText(string(cp))
	p.StartIndex = endIndex
}

/**
* Checks if the current tag is a void element. Override this if you want
* to specify your own additional void elements.
 */
func (p *Parser) isVoidElement(name string) bool {
	if !p.htmlMode {
		return false
	}

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

/** @internal */
func (p *Parser) OnOpenTagName(start, endIndex int) {
	p.EndIndex = endIndex

	name := p.getSlice(start, endIndex)

	if p.lowerCaseTagNames {
		name = bytes.ToLower(name)
	}

	p.emitOpenTag(string(name))
}

func (p *Parser) emitOpenTag(name string) {
	p.openTagStart = p.StartIndex
	p.tagname = name

	impliesClose, has := openImpliesClose[name]

	if p.htmlMode && has {
		for len(p.stack) > 0 && impliesClose(p.stack[len(p.stack)-1]) {
			element := p.stack[len(p.stack)-1]
			p.stack = p.stack[:len(p.stack)-1]
			p.cbs.OnCloseTag(element, true)
		}
	}
	if !p.isVoidElement(name) {
		p.stack = append(p.stack, name)

		if p.htmlMode {
			if isForeignContextElements(name) {
				p.foreignContext = append(p.foreignContext, true)
			} else if isHtmlIntegrationElements(name) {
				p.foreignContext = append(p.foreignContext, false)
			}
		}
	}
	p.cbs.OnOpenTagName(name)
	p.attribs = make([]*Attribute, 0)
}

func (p *Parser) endOpenTag(isImplied bool) {
	p.StartIndex = p.openTagStart

	if p.attribs != nil {
		p.cbs.OnOpenTag(p.tagname, p.attribs, isImplied)
		p.attribs = nil
	}
	if p.isVoidElement(p.tagname) {
		p.cbs.OnCloseTag(p.tagname, true)
	}

	p.tagname = ""
}

/** @internal */
func (p *Parser) OnOpenTagEnd(endIndex int) {
	p.EndIndex = endIndex
	p.endOpenTag(false)

	// Set `startIndex` for next node
	p.StartIndex = endIndex + 1
}

/** @internal */
func (p *Parser) OnCloseTag(start, endIndex int) {
	p.EndIndex = endIndex

	name := string(p.getSlice(start, endIndex))

	if p.lowerCaseTagNames {
		name = strings.ToLower(name)
	}

	if p.htmlMode && (isForeignContextElements(name) || isHtmlIntegrationElements(name)) {
		p.foreignContext = p.foreignContext[:len(p.foreignContext)-1]
	}

	if !p.isVoidElement(name) {
		pos := lastIndex(p.stack, name)
		if pos != -1 {
			for index := len(p.stack) - 1; index >= pos; index-- {
				element := p.stack[len(p.stack)-1]
				p.stack = p.stack[:len(p.stack)-1]
				// We know the stack has sufficient elements.
				p.cbs.OnCloseTag(element, index != pos)
			}
		} else if p.htmlMode && name == "p" {
			// Implicit open before close
			p.emitOpenTag("p")
			p.closeCurrentTag(true)
		}
	} else if p.htmlMode && name == "br" {
		// We can't use `emitOpenTag` for implicit open, as `br` would be implicitly closed.
		p.cbs.OnOpenTagName("br")
		p.cbs.OnOpenTag("br", nil, true)
		p.cbs.OnCloseTag("br", false)
	}

	// Set `startIndex` for next node
	p.StartIndex = endIndex + 1
}

/** @internal */
func (p *Parser) OnSelfClosingTag(endIndex int) {
	p.EndIndex = endIndex
	if p.recognizeSelfClosing || p.foreignContext[0] {
		p.closeCurrentTag(false)

		// Set `startIndex` for next node
		p.StartIndex = endIndex + 1
	} else {
		// Ignore the fact that the tag is self-closing.
		p.OnOpenTagEnd(endIndex)
	}
}

func (p *Parser) closeCurrentTag(isOpenImplied bool) {
	name := p.tagname
	p.endOpenTag(isOpenImplied)

	// Self-closing tags will be on the top of the stack
	if len(p.stack) > 0 && p.stack[len(p.stack)-1] == name {
		// If the opening tag isn't implied, the closing tag has to be implied.
		p.cbs.OnCloseTag(name, !isOpenImplied)
		p.stack = p.stack[:len(p.stack)-1]
	}
}

/** @internal */
func (p *Parser) OnAttribName(start, endIndex int) {
	p.StartIndex = start
	name := string(p.getSlice(start, endIndex))

	if p.lowerCaseAttributeNames {
		p.attribname = strings.ToLower(name)
	} else {
		p.attribname = name
	}
}

/** @internal */
func (p *Parser) OnAttribData(start, endIndex int) {
	p.attribvalue = append(p.attribvalue, p.getSlice(start, endIndex)...)
}

/** @internal */
func (p *Parser) OnAttribEntity(cp rune) {
	p.attribvalue = utf8.AppendRune(p.attribvalue, cp)
}

/** @internal */
func (p *Parser) OnAttribEnd(quote QuoteType, endIndex int) {
	p.EndIndex = endIndex
	attribvalue := string(p.attribvalue)

	p.cbs.OnAttribute(
		p.attribname,
		attribvalue,
		quote,
	)

	if p.attribs != nil && slices.IndexFunc(p.attribs, func(a *Attribute) bool { return a.Name == p.attribname }) == -1 {
		p.attribs = append(p.attribs, &Attribute{Name: p.attribname, Value: attribvalue})
	}

	p.attribvalue = nil
}

func (p *Parser) getInstructionName(value []byte) []byte {
	index := bytes.IndexAny(value, "/ \t\n\r\v\f")
	var name []byte
	if index < 0 {
		name = value
	} else {
		name = value[:index]
	}

	if p.lowerCaseTagNames {
		name = bytes.ToLower(name)
	}

	return name
}

/** @internal */
func (p *Parser) OnDeclaration(start, endIndex int) {
	p.EndIndex = endIndex
	value := p.getSlice(start, endIndex)

	name := p.getInstructionName(value)
	p.cbs.OnProcessingInstruction(fmt.Sprintf("!%s", name), fmt.Sprintf("!%s", value))

	// Set `startIndex` for next node
	p.StartIndex = endIndex + 1
}

/** @internal */
func (p *Parser) OnProcessingInstruction(start, endIndex int) {
	p.EndIndex = endIndex
	value := p.getSlice(start, endIndex)

	name := p.getInstructionName(value)
	p.cbs.OnProcessingInstruction(fmt.Sprintf("?%s", name), fmt.Sprintf("?%s", value))

	// Set `startIndex` for next node
	p.StartIndex = endIndex + 1
}

/** @internal */
func (p *Parser) OnComment(start, endIndex, offset int) {
	p.EndIndex = endIndex

	p.cbs.OnComment(string(p.getSlice(start, endIndex-offset)))
	p.cbs.OnCommentEnd()

	// Set `startIndex` for next node
	p.StartIndex = endIndex + 1
}

/** @internal */
func (p *Parser) OnCData(start, endIndex, offset int) {
	p.EndIndex = endIndex
	value := p.getSlice(start, endIndex-offset)

	if !p.htmlMode || p.recognizeCDATA {
		p.cbs.OnCDataStart()
		p.cbs.OnText(string(value))
		p.cbs.OnCDataEnd()
	} else {
		p.cbs.OnComment(fmt.Sprintf("[CDATA[%s]]", value))
		p.cbs.OnCommentEnd()
	}

	// Set `startIndex` for next node
	p.StartIndex = endIndex + 1
}

/** @internal */
func (p *Parser) OnEnd() {
	// Set the end index for all remaining tags
	p.EndIndex = p.StartIndex
	for index := len(p.stack) - 1; index >= 0; index-- {
		p.cbs.OnCloseTag(p.stack[index], true)
	}

	p.cbs.OnEnd()
}

/**
* Resets the parser to a blank state, ready to parse a new HTML document
 */
func (p *Parser) Reset() {
	p.cbs.OnReset()
	p.tokenizer.Reset()
	p.tagname = ""
	p.attribname = ""
	p.attribs = nil
	p.stack = make([]string, 0, 10)
	p.StartIndex = 0
	p.EndIndex = 0
	p.cbs.OnParserInit(p)
	p.buffers = make([][]byte, 0, 10)
	p.foreignContext = []bool{!p.htmlMode}
	p.bufferOffset = 0
	p.writeIndex = 0
	p.ended = false
}

/**
* Resets the parser, then parses a complete document and
* pushes it to the handler.
*
* @param data Document to parse.
 */
func (p *Parser) ParseComplete(data []byte) {
	p.Reset()
	p.End(data)
}

func (p *Parser) getSlice(start, end int) []byte {
	for start-p.bufferOffset >= len(p.buffers[0]) {
		p.shiftBuffer()
	}

	actualEnd := end - p.bufferOffset
	if actualEnd > len(p.buffers[0]) {
		actualEnd = len(p.buffers[0])
	}
	slice := make([]byte, 0, end-start)
	slice = append(slice, p.buffers[0][start-p.bufferOffset:actualEnd]...)

	for end-p.bufferOffset > len(p.buffers[0]) {
		p.shiftBuffer()
		actualEnd = end - p.bufferOffset
		if actualEnd > len(p.buffers[0]) {
			actualEnd = len(p.buffers[0])
		}
		slice = append(slice, p.buffers[0][0:actualEnd]...)
	}

	return slice
}

func (p *Parser) shiftBuffer() {
	p.bufferOffset += len(p.buffers[0])
	p.writeIndex--
	p.buffers[0] = nil
	p.buffers = p.buffers[1:]
}

/**
* Parses a chunk of data and calls the corresponding callbacks.
*
* @param chunk Chunk to parse.
 */
func (p *Parser) Write(chunk []byte) {
	if p.ended {
		p.cbs.OnError(errors.New(".write() after done"))
		return
	}

	p.buffers = append(p.buffers, chunk)
	if p.tokenizer.running {
		p.tokenizer.Write(chunk)
		p.writeIndex++
	}
}

/**
* Parses the end of the buffer and clears the stack, calls onend.
*
* @param chunk Optional final chunk to parse.
 */
func (p *Parser) End(chunk []byte) {
	if p.ended {
		p.cbs.OnError(errors.New(".end() after done"))
		return
	}

	if chunk != nil {
		p.Write(chunk)
	}
	p.ended = true
	p.tokenizer.End()
}

/**
* Pauses parsing. The parser won't emit events until `resume` is called.
 */
func (p *Parser) Pause() {
	p.tokenizer.Pause()
}

/**
* Resumes parsing after `pause` was called.
 */
func (p *Parser) Resume() {
	p.tokenizer.Resume()

	for p.tokenizer.running && p.writeIndex < len(p.buffers) {
		p.tokenizer.Write(p.buffers[p.writeIndex])
		p.writeIndex++
	}

	if p.ended {
		p.tokenizer.End()
	}
}
