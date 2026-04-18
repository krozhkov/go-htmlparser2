package parser

import (
	"bytes"

	"github.com/krozhkov/go-htmlparser2/entities"
)

const (
	Tab                  = 0x9  // "\t"
	NewLine              = 0xa  // "\n"
	FormFeed             = 0xc  // "\f"
	CarriageReturn       = 0xd  // "\r"
	Space                = 0x20 // " "
	ExclamationMark      = 0x21 // "!"
	Number               = 0x23 // "#"
	Amp                  = 0x26 // "&"
	SingleQuote          = 0x27 // "'"
	DoubleQuote          = 0x22 // '"'
	Dash                 = 0x2d // "-"
	Slash                = 0x2f // "/"
	Zero                 = 0x30 // "0"
	Nine                 = 0x39 // "9"
	Semi                 = 0x3b // ";"
	Lt                   = 0x3c // "<"
	Eq                   = 0x3d // "="
	Gt                   = 0x3e // ">"
	Questionmark         = 0x3f // "?"
	UpperA               = 0x41 // "A"
	LowerA               = 0x61 // "a"
	UpperF               = 0x46 // "F"
	LowerF               = 0x66 // "f"
	UpperZ               = 0x5a // "Z"
	LowerZ               = 0x7a // "z"
	LowerX               = 0x78 // "x"
	OpeningSquareBracket = 0x5b // "["
)

/** All the states the tokenizer can be in. */
type State uint32

const (
	StateText          State = iota + 1
	StateBeforeTagName       // After <
	StateInTagName
	StateInSelfClosingTag
	StateBeforeClosingTagName
	StateInClosingTagName
	StateAfterClosingTagName

	// Attributes
	StateBeforeAttributeName
	StateInAttributeName
	StateAfterAttributeName
	StateBeforeAttributeValue
	StateInAttributeValueDq // "
	StateInAttributeValueSq // '
	StateInAttributeValueNq

	// Declarations
	StateBeforeDeclaration // !
	StateInDeclaration

	// Processing instructions
	StateInProcessingInstruction // ?

	// Comments & CDATA
	StateBeforeComment
	StateCDATASequence
	StateInSpecialComment
	StateInCommentLike

	// Special tags
	StateBeforeSpecialS // Decide if we deal with `<script` or `<style`
	StateBeforeSpecialT // Decide if we deal with `<title` or `<textarea`
	StateSpecialStartSequence
	StateInSpecialTag

	StateInEntity
)

func isWhitespace(c byte) bool {
	return c == Space ||
		c == NewLine ||
		c == Tab ||
		c == FormFeed ||
		c == CarriageReturn
}

func isEndOfTagSection(c byte) bool {
	return c == Slash || c == Gt || isWhitespace(c)
}

func isASCIIAlpha(c byte) bool {
	return (c >= LowerA && c <= LowerZ) ||
		(c >= UpperA && c <= UpperZ)
}

type QuoteType int

const (
	QuoteTypeNoValue QuoteType = iota
	QuoteTypeUnquoted
	QuoteTypeSingle
	QuoteTypeDouble
)

type Callbacks interface {
	OnAttribData(start, endIndex int)
	OnAttribEntity(codepoint rune)
	OnAttribEnd(quote QuoteType, endIndex int)
	OnAttribName(start, endIndex int)
	OnCData(start, endIndex, endOffset int)
	OnCloseTag(start, endIndex int)
	OnComment(start, endIndex, endOffset int)
	OnDeclaration(start, endIndex int)
	OnEnd()
	OnOpenTagEnd(endIndex int)
	OnOpenTagName(start, endIndex int)
	OnProcessingInstruction(start, endIndex int)
	OnSelfClosingTag(endIndex int)
	OnText(start, endIndex int)
	OnTextEntity(codepoint rune, endIndex int)
}

/**
 * Sequences used to match longer strings.
 *
 * We don't have `Script`, `Style`, or `Title` here. Instead, we re-use the *End
 * sequences with an increased offset.
 */
var Sequences = struct {
	Cdata       []byte
	CdataEnd    []byte
	CommentEnd  []byte
	ScriptEnd   []byte
	StyleEnd    []byte
	TitleEnd    []byte
	TextareaEnd []byte
	XmpEnd      []byte
}{
	Cdata:       []byte{0x43, 0x44, 0x41, 0x54, 0x41, 0x5b},                         // CDATA[
	CdataEnd:    []byte{0x5d, 0x5d, 0x3e},                                           // ]]>
	CommentEnd:  []byte{0x2d, 0x2d, 0x3e},                                           // `-->`
	ScriptEnd:   []byte{0x3c, 0x2f, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74},             // `</script`
	StyleEnd:    []byte{0x3c, 0x2f, 0x73, 0x74, 0x79, 0x6c, 0x65},                   // `</style`
	TitleEnd:    []byte{0x3c, 0x2f, 0x74, 0x69, 0x74, 0x6c, 0x65},                   // `</title`
	TextareaEnd: []byte{0x3c, 0x2f, 0x74, 0x65, 0x78, 0x74, 0x61, 0x72, 0x65, 0x61}, // `</textarea`
	XmpEnd:      []byte{0x3c, 0x2f, 0x78, 0x6d, 0x70},                               // `</xmp`
}

type TokenizerOptions struct {
	XmlMode        bool
	DecodeEntities bool
}

type Tokenizer struct {
	/** The current state the tokenizer is in. */
	state State
	/** The read buffer. */
	buffer []byte
	/** The beginning of the section that is currently being read. */
	sectionStart int
	/** The index within the buffer that we are currently looking at. */
	index int
	/** The start of the last entity. */
	entityStart int
	/** Some behavior, eg. when decoding entities, is done while we are in another state. This keeps track of the other state type. */
	baseState State
	/** For special parsing behavior inside of script and style tags. */
	isSpecial bool
	/** Indicates whether the tokenizer has been paused. */
	running bool
	/** The offset of the current buffer. */
	offset int

	currentSequence []byte
	sequenceIndex   int

	xmlMode        bool
	decodeEntities bool

	entityDecoder *entities.EntityDecoder

	cbs Callbacks
}

func NewTokenizer(options TokenizerOptions, cbs Callbacks) *Tokenizer {
	t := &Tokenizer{
		state:          StateText,
		sectionStart:   0,
		index:          0,
		entityStart:    0,
		baseState:      StateText,
		isSpecial:      false,
		running:        true,
		offset:         0,
		sequenceIndex:  0,
		xmlMode:        options.XmlMode,
		decodeEntities: options.DecodeEntities,
		cbs:            cbs,
	}

	var decodeTree []uint16
	if options.XmlMode {
		decodeTree = entities.XmlDecodeTree
	} else {
		decodeTree = entities.HtmlDecodeTree
	}

	entityDecoder := entities.NewEntityDecoder(
		decodeTree,
		t.emitCodePoint,
		nil,
	)

	t.entityDecoder = entityDecoder

	return t
}

func (t *Tokenizer) Reset() {
	t.state = StateText
	t.buffer = nil
	t.sectionStart = 0
	t.index = 0
	t.baseState = StateText
	t.currentSequence = nil
	t.running = true
	t.offset = 0
}

func (t *Tokenizer) Write(chunk []byte) {
	t.offset += len(t.buffer)
	t.buffer = chunk
	t.parse()
}

func (t *Tokenizer) End() {
	if t.running {
		t.finish()
	}
}

func (t *Tokenizer) Pause() {
	t.running = false
}

func (t *Tokenizer) Resume() {
	t.running = true
	if t.index < len(t.buffer)+t.offset {
		t.parse()
	}
}

func (t *Tokenizer) stateText(c byte) {
	if c == Lt || (!t.decodeEntities && t.fastForwardTo(Lt)) {
		if t.index > t.sectionStart {
			t.cbs.OnText(t.sectionStart, t.index)
		}
		t.state = StateBeforeTagName
		t.sectionStart = t.index
	} else if t.decodeEntities && c == Amp {
		t.startEntity()
	}
}

func (t *Tokenizer) stateSpecialStartSequence(c byte) {
	isEnd := t.sequenceIndex == len(t.currentSequence)
	var isMatch bool
	if isEnd {
		// If we are at the end of the sequence, make sure the tag name has ended
		isMatch = isEndOfTagSection(c)
	} else {
		// Otherwise, do a case-insensitive comparison
		isMatch = (c | 0x20) == t.currentSequence[t.sequenceIndex]
	}

	if !isMatch {
		t.isSpecial = false
	} else if !isEnd {
		t.sequenceIndex++
		return
	}

	t.sequenceIndex = 0
	t.state = StateInTagName
	t.stateInTagName(c)
}

/** Look for an end tag. For <title> tags, also decode entities. */
func (t *Tokenizer) stateInSpecialTag(c byte) {
	if t.sequenceIndex == len(t.currentSequence) {
		if c == Gt || isWhitespace(c) {
			endOfText := t.index - len(t.currentSequence)

			if t.sectionStart < endOfText {
				// Spoof the index so that reported locations match up.
				actualIndex := t.index
				t.index = endOfText
				t.cbs.OnText(t.sectionStart, endOfText)
				t.index = actualIndex
			}

			t.isSpecial = false
			t.sectionStart = endOfText + 2 // Skip over the `</`
			t.stateInClosingTagName(c)
			return // We are done; skip the rest of the function.
		}

		t.sequenceIndex = 0
	}

	if (c | 0x20) == t.currentSequence[t.sequenceIndex] {
		t.sequenceIndex += 1
	} else if t.sequenceIndex == 0 {
		if bytes.Equal(t.currentSequence, Sequences.TitleEnd) {
			// We have to parse entities in <title> tags.
			if t.decodeEntities && c == Amp {
				t.startEntity()
			}
		} else if t.fastForwardTo(Lt) {
			// Outside of <title> tags, we can fast-forward.
			t.sequenceIndex = 1
		}
	} else {
		// If we see a `<`, set the sequence index to 1; useful for eg. `<</script>`.
		if c == Lt {
			t.sequenceIndex = 1
		} else {
			t.sequenceIndex = 0
		}
	}
}

func (t *Tokenizer) stateCDATASequence(c byte) {
	if c == Sequences.Cdata[t.sequenceIndex] {
		t.sequenceIndex++
		if t.sequenceIndex == len(Sequences.Cdata) {
			t.state = StateInCommentLike
			t.currentSequence = Sequences.CdataEnd
			t.sequenceIndex = 0
			t.sectionStart = t.index + 1
		}
	} else {
		t.sequenceIndex = 0
		t.state = StateInDeclaration
		t.stateInDeclaration(c) // Reconsume the character
	}
}

/**
* When we wait for one specific character, we can speed things up
* by skipping through the buffer until we find it.
*
* @returns Whether the character was found.
 */
func (t *Tokenizer) fastForwardTo(c byte) bool {
	for {
		t.index++
		if t.index >= len(t.buffer)+t.offset {
			break
		}

		if t.buffer[t.index-t.offset] == c {
			return true
		}
	}

	/*
	* We increment the index at the end of the `parse` loop,
	* so set it to `buffer.length - 1` here.
	*
	* TODO: Refactor `parse` to increment index before calling states.
	 */
	t.index = len(t.buffer) + t.offset - 1

	return false
}

/**
* Comments and CDATA end with `-->` and `]]>`.
*
* Their common qualities are:
* - Their end sequences have a distinct character they start with.
* - That character is then repeated, so we have to check multiple repeats.
* - All characters but the start character of the sequence can be skipped.
 */
func (t *Tokenizer) stateInCommentLike(c byte) {
	if c == t.currentSequence[t.sequenceIndex] {
		t.sequenceIndex++
		if t.sequenceIndex == len(t.currentSequence) {
			if bytes.Equal(t.currentSequence, Sequences.CdataEnd) {
				t.cbs.OnCData(t.sectionStart, t.index, 2)
			} else {
				t.cbs.OnComment(t.sectionStart, t.index, 2)
			}

			t.sequenceIndex = 0
			t.sectionStart = t.index + 1
			t.state = StateText
		}
	} else if t.sequenceIndex == 0 {
		// Fast-forward to the first character of the sequence
		if t.fastForwardTo(t.currentSequence[0]) {
			t.sequenceIndex = 1
		}
	} else if c != t.currentSequence[t.sequenceIndex-1] {
		// Allow long sequences, eg. --->, ]]]>
		t.sequenceIndex = 0
	}
}

/**
* HTML only allows ASCII alpha characters (a-z and A-Z) at the beginning of a tag name.
*
* XML allows a lot more characters here (@see https://www.w3.org/TR/REC-xml/#NT-NameStartChar).
* We allow anything that wouldn't end the tag.
 */
func (t *Tokenizer) isTagStartChar(c byte) bool {
	if t.xmlMode {
		return !isEndOfTagSection(c)
	} else {
		return isASCIIAlpha(c)
	}
}

func (t *Tokenizer) startSpecial(sequence []byte, offset int) {
	t.isSpecial = true
	t.currentSequence = sequence
	t.sequenceIndex = offset
	t.state = StateSpecialStartSequence
}

func (t *Tokenizer) stateBeforeTagName(c byte) {
	if c == ExclamationMark {
		t.state = StateBeforeDeclaration
		t.sectionStart = t.index + 1
	} else if c == Questionmark {
		t.state = StateInProcessingInstruction
		t.sectionStart = t.index + 1
	} else if t.isTagStartChar(c) {
		lower := c | 0x20
		t.sectionStart = t.index
		if t.xmlMode {
			t.state = StateInTagName
		} else if lower == Sequences.ScriptEnd[2] {
			t.state = StateBeforeSpecialS
		} else if lower == Sequences.TitleEnd[2] ||
			lower == Sequences.XmpEnd[2] {
			t.state = StateBeforeSpecialT
		} else {
			t.state = StateInTagName
		}
	} else if c == Slash {
		t.state = StateBeforeClosingTagName
	} else {
		t.state = StateText
		t.stateText(c)
	}
}

func (t *Tokenizer) stateInTagName(c byte) {
	if isEndOfTagSection(c) {
		t.cbs.OnOpenTagName(t.sectionStart, t.index)
		t.sectionStart = -1
		t.state = StateBeforeAttributeName
		t.stateBeforeAttributeName(c)
	}
}

func (t *Tokenizer) stateBeforeClosingTagName(c byte) {
	if isWhitespace(c) {
		// Ignore
	} else if c == Gt {
		t.state = StateText
	} else {
		if t.isTagStartChar(c) {
			t.state = StateInClosingTagName
		} else {
			t.state = StateInSpecialComment
		}
		t.sectionStart = t.index
	}
}

func (t *Tokenizer) stateInClosingTagName(c byte) {
	if c == Gt || isWhitespace(c) {
		t.cbs.OnCloseTag(t.sectionStart, t.index)
		t.sectionStart = -1
		t.state = StateAfterClosingTagName
		t.stateAfterClosingTagName(c)
	}
}

func (t *Tokenizer) stateAfterClosingTagName(c byte) {
	// Skip everything until ">"
	if c == Gt || t.fastForwardTo(Gt) {
		t.state = StateText
		t.sectionStart = t.index + 1
	}
}

func (t *Tokenizer) stateBeforeAttributeName(c byte) {
	if c == Gt {
		t.cbs.OnOpenTagEnd(t.index)
		if t.isSpecial {
			t.state = StateInSpecialTag
			t.sequenceIndex = 0
		} else {
			t.state = StateText
		}
		t.sectionStart = t.index + 1
	} else if c == Slash {
		t.state = StateInSelfClosingTag
	} else if !isWhitespace(c) {
		t.state = StateInAttributeName
		t.sectionStart = t.index
	}
}

func (t *Tokenizer) stateInSelfClosingTag(c byte) {
	if c == Gt {
		t.cbs.OnSelfClosingTag(t.index)
		t.state = StateText
		t.sectionStart = t.index + 1
		t.isSpecial = false // Reset special state, in case of self-closing special tags
	} else if !isWhitespace(c) {
		t.state = StateBeforeAttributeName
		t.stateBeforeAttributeName(c)
	}
}

func (t *Tokenizer) stateInAttributeName(c byte) {
	if c == Eq || isEndOfTagSection(c) {
		t.cbs.OnAttribName(t.sectionStart, t.index)
		t.sectionStart = t.index
		t.state = StateAfterAttributeName
		t.stateAfterAttributeName(c)
	}
}

func (t *Tokenizer) stateAfterAttributeName(c byte) {
	if c == Eq {
		t.state = StateBeforeAttributeValue
	} else if c == Slash || c == Gt {
		t.cbs.OnAttribEnd(QuoteTypeNoValue, t.sectionStart)
		t.sectionStart = -1
		t.state = StateBeforeAttributeName
		t.stateBeforeAttributeName(c)
	} else if !isWhitespace(c) {
		t.cbs.OnAttribEnd(QuoteTypeNoValue, t.sectionStart)
		t.state = StateInAttributeName
		t.sectionStart = t.index
	}
}

func (t *Tokenizer) stateBeforeAttributeValue(c byte) {
	if c == DoubleQuote {
		t.state = StateInAttributeValueDq
		t.sectionStart = t.index + 1
	} else if c == SingleQuote {
		t.state = StateInAttributeValueSq
		t.sectionStart = t.index + 1
	} else if !isWhitespace(c) {
		t.sectionStart = t.index
		t.state = StateInAttributeValueNq
		t.stateInAttributeValueNoQuotes(c) // Reconsume token
	}
}

func (t *Tokenizer) handleInAttributeValue(c byte, quote byte) {
	if c == quote || (!t.decodeEntities && t.fastForwardTo(quote)) {
		t.cbs.OnAttribData(t.sectionStart, t.index)
		t.sectionStart = -1
		var quoteType QuoteType
		if quote == DoubleQuote {
			quoteType = QuoteTypeDouble
		} else {
			quoteType = QuoteTypeSingle
		}
		t.cbs.OnAttribEnd(quoteType, t.index+1)
		t.state = StateBeforeAttributeName
	} else if t.decodeEntities && c == Amp {
		t.startEntity()
	}
}

func (t *Tokenizer) stateInAttributeValueDoubleQuotes(c byte) {
	t.handleInAttributeValue(c, DoubleQuote)
}

func (t *Tokenizer) stateInAttributeValueSingleQuotes(c byte) {
	t.handleInAttributeValue(c, SingleQuote)
}

func (t *Tokenizer) stateInAttributeValueNoQuotes(c byte) {
	if isWhitespace(c) || c == Gt {
		t.cbs.OnAttribData(t.sectionStart, t.index)
		t.sectionStart = -1
		t.cbs.OnAttribEnd(QuoteTypeUnquoted, t.index)
		t.state = StateBeforeAttributeName
		t.stateBeforeAttributeName(c)
	} else if t.decodeEntities && c == Amp {
		t.startEntity()
	}
}

func (t *Tokenizer) stateBeforeDeclaration(c byte) {
	if c == OpeningSquareBracket {
		t.state = StateCDATASequence
		t.sequenceIndex = 0
	} else {
		if c == Dash {
			t.state = StateBeforeComment
		} else {
			t.state = StateInDeclaration
		}
	}
}

func (t *Tokenizer) stateInDeclaration(c byte) {
	if c == Gt || t.fastForwardTo(Gt) {
		t.cbs.OnDeclaration(t.sectionStart, t.index)
		t.state = StateText
		t.sectionStart = t.index + 1
	}
}

func (t *Tokenizer) stateInProcessingInstruction(c byte) {
	if c == Gt || t.fastForwardTo(Gt) {
		t.cbs.OnProcessingInstruction(t.sectionStart, t.index)
		t.state = StateText
		t.sectionStart = t.index + 1
	}
}

func (t *Tokenizer) stateBeforeComment(c byte) {
	if c == Dash {
		t.state = StateInCommentLike
		t.currentSequence = Sequences.CommentEnd
		// Allow short comments (eg. <!-->)
		t.sequenceIndex = 2
		t.sectionStart = t.index + 1
	} else {
		t.state = StateInDeclaration
	}
}

func (t *Tokenizer) stateInSpecialComment(c byte) {
	if c == Gt || t.fastForwardTo(Gt) {
		t.cbs.OnComment(t.sectionStart, t.index, 0)
		t.state = StateText
		t.sectionStart = t.index + 1
	}
}

func (t *Tokenizer) stateBeforeSpecialS(c byte) {
	lower := c | 0x20
	if lower == Sequences.ScriptEnd[3] {
		t.startSpecial(Sequences.ScriptEnd, 4)
	} else if lower == Sequences.StyleEnd[3] {
		t.startSpecial(Sequences.StyleEnd, 4)
	} else {
		t.state = StateInTagName
		t.stateInTagName(c) // Consume the token again
	}
}

func (t *Tokenizer) stateBeforeSpecialT(c byte) {
	lower := c | 0x20
	switch lower {
	case Sequences.TitleEnd[3]:
		{
			t.startSpecial(Sequences.TitleEnd, 4)
			break
		}
	case Sequences.TextareaEnd[3]:
		{
			t.startSpecial(Sequences.TextareaEnd, 4)
			break
		}
	case Sequences.XmpEnd[3]:
		{
			t.startSpecial(Sequences.XmpEnd, 4)
			break
		}
	default:
		{
			t.state = StateInTagName
			t.stateInTagName(c) // Consume the token again
		}
	}
}

func (t *Tokenizer) startEntity() {
	var decodeMode entities.DecodingMode
	if t.xmlMode {
		decodeMode = entities.DecodingModeStrict
	} else if t.baseState == StateText || t.baseState == StateInSpecialTag {
		decodeMode = entities.DecodingModeLegacy
	} else {
		decodeMode = entities.DecodingModeAttribute
	}

	t.baseState = t.state
	t.state = StateInEntity
	t.entityStart = t.index
	t.entityDecoder.StartEntity(decodeMode)
}

func (t *Tokenizer) stateInEntity() {
	length := t.entityDecoder.Write(
		t.buffer,
		t.index-t.offset,
	)

	// If `length` is positive, we are done with the entity.
	if length >= 0 {
		t.state = t.baseState

		if length == 0 {
			t.index = t.entityStart
		}
	} else {
		// Mark buffer as consumed.
		t.index = t.offset + len(t.buffer) - 1
	}
}

/**
* Remove data that has already been consumed from the buffer.
 */
func (t *Tokenizer) cleanup() {
	// If we are inside of text or attributes, emit what we already have.
	if t.running && t.sectionStart != t.index {
		if t.state == StateText ||
			(t.state == StateInSpecialTag && t.sequenceIndex == 0) {
			t.cbs.OnText(t.sectionStart, t.index)
			t.sectionStart = t.index
		} else if t.state == StateInAttributeValueDq ||
			t.state == StateInAttributeValueSq ||
			t.state == StateInAttributeValueNq {
			t.cbs.OnAttribData(t.sectionStart, t.index)
			t.sectionStart = t.index
		}
	}
}

func (t *Tokenizer) shouldContinue() bool {
	return t.index < (len(t.buffer)+t.offset) && t.running
}

/**
 * Iterates through the buffer, calling the function corresponding to the current state.
 *
 * States that are more likely to be hit are higher up, as a performance improvement.
 */
func (t *Tokenizer) parse() {
	for t.shouldContinue() {
		if t.index >= t.offset {
			c := t.buffer[t.index-t.offset]
			switch t.state {
			case StateText:
				{
					t.stateText(c)
					break
				}
			case StateSpecialStartSequence:
				{
					t.stateSpecialStartSequence(c)
					break
				}
			case StateInSpecialTag:
				{
					t.stateInSpecialTag(c)
					break
				}
			case StateCDATASequence:
				{
					t.stateCDATASequence(c)
					break
				}
			case StateInAttributeValueDq:
				{
					t.stateInAttributeValueDoubleQuotes(c)
					break
				}
			case StateInAttributeName:
				{
					t.stateInAttributeName(c)
					break
				}
			case StateInCommentLike:
				{
					t.stateInCommentLike(c)
					break
				}
			case StateInSpecialComment:
				{
					t.stateInSpecialComment(c)
					break
				}
			case StateBeforeAttributeName:
				{
					t.stateBeforeAttributeName(c)
					break
				}
			case StateInTagName:
				{
					t.stateInTagName(c)
					break
				}
			case StateInClosingTagName:
				{
					t.stateInClosingTagName(c)
					break
				}
			case StateBeforeTagName:
				{
					t.stateBeforeTagName(c)
					break
				}
			case StateAfterAttributeName:
				{
					t.stateAfterAttributeName(c)
					break
				}
			case StateInAttributeValueSq:
				{
					t.stateInAttributeValueSingleQuotes(c)
					break
				}
			case StateBeforeAttributeValue:
				{
					t.stateBeforeAttributeValue(c)
					break
				}
			case StateBeforeClosingTagName:
				{
					t.stateBeforeClosingTagName(c)
					break
				}
			case StateAfterClosingTagName:
				{
					t.stateAfterClosingTagName(c)
					break
				}
			case StateBeforeSpecialS:
				{
					t.stateBeforeSpecialS(c)
					break
				}
			case StateBeforeSpecialT:
				{
					t.stateBeforeSpecialT(c)
					break
				}
			case StateInAttributeValueNq:
				{
					t.stateInAttributeValueNoQuotes(c)
					break
				}
			case StateInSelfClosingTag:
				{
					t.stateInSelfClosingTag(c)
					break
				}
			case StateInDeclaration:
				{
					t.stateInDeclaration(c)
					break
				}
			case StateBeforeDeclaration:
				{
					t.stateBeforeDeclaration(c)
					break
				}
			case StateBeforeComment:
				{
					t.stateBeforeComment(c)
					break
				}
			case StateInProcessingInstruction:
				{
					t.stateInProcessingInstruction(c)
					break
				}
			case StateInEntity:
				{
					t.stateInEntity()
					break
				}
			}
		}
		t.index++
	}
	t.cleanup()
}

func (t *Tokenizer) finish() {
	if t.state == StateInEntity {
		t.entityDecoder.End()
		t.state = t.baseState
	}

	t.handleTrailingData()

	t.cbs.OnEnd()
}

/** Handle any trailing data. */
func (t *Tokenizer) handleTrailingData() {
	endIndex := len(t.buffer) + t.offset

	// If there is no remaining data, we are done.
	if t.sectionStart >= endIndex {
		return
	}

	if t.state == StateInCommentLike {
		if bytes.Equal(t.currentSequence, Sequences.CdataEnd) {
			t.cbs.OnCData(t.sectionStart, endIndex, 0)
		} else {
			t.cbs.OnComment(t.sectionStart, endIndex, 0)
		}
	} else if t.state == StateInTagName ||
		t.state == StateBeforeAttributeName ||
		t.state == StateBeforeAttributeValue ||
		t.state == StateAfterAttributeName ||
		t.state == StateInAttributeName ||
		t.state == StateInAttributeValueSq ||
		t.state == StateInAttributeValueDq ||
		t.state == StateInAttributeValueNq ||
		t.state == StateInClosingTagName {
		/*
		* If we are currently in an opening or closing tag, us not calling the
		* respective callback signals that the tag should be ignored.
		 */
	} else {
		t.cbs.OnText(t.sectionStart, endIndex)
	}
}

func (t *Tokenizer) emitCodePoint(cp rune, consumed int) {
	if t.baseState != StateText &&
		t.baseState != StateInSpecialTag {
		if t.sectionStart < t.entityStart {
			t.cbs.OnAttribData(t.sectionStart, t.entityStart)
		}
		t.sectionStart = t.entityStart + consumed
		t.index = t.sectionStart - 1

		t.cbs.OnAttribEntity(cp)
	} else {
		if t.sectionStart < t.entityStart {
			t.cbs.OnText(t.sectionStart, t.entityStart)
		}
		t.sectionStart = t.entityStart + consumed
		t.index = t.sectionStart - 1

		t.cbs.OnTextEntity(cp, t.sectionStart)
	}
}
