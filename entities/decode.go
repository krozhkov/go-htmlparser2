package entities

import (
	"unicode/utf8"
)

const (
	NUM     = 35  // "#"
	SEMI    = 59  // ";"
	EQUALS  = 61  // "="
	ZERO    = 48  // "0"
	NINE    = 57  // "9"
	LOWER_A = 97  // "a"
	LOWER_F = 102 // "f"
	LOWER_X = 120 // "x"
	LOWER_Z = 122 // "z"
	UPPER_A = 65  // "A"
	UPPER_F = 70  // "F"
	UPPER_Z = 90  // "Z"
)

/** Bit that needs to be set to convert an upper case ASCII character to lower case */
const TO_LOWER_BIT = 0b10_0000

func isNumber(code byte) bool {
	return code >= ZERO && code <= NINE
}

func isHexadecimalCharacter(code byte) bool {
	return (code >= UPPER_A && code <= UPPER_F) ||
		(code >= LOWER_A && code <= LOWER_F)
}

func isAsciiAlphaNumeric(code byte) bool {
	return (code >= UPPER_A && code <= UPPER_Z) ||
		(code >= LOWER_A && code <= LOWER_Z) ||
		isNumber(code)
}

/**
 * Checks if the given character is a valid end character for an entity in an attribute.
 *
 * Attribute values that aren't terminated properly aren't parsed, and shouldn't lead to a parser error.
 * See the example in https://html.spec.whatwg.org/multipage/parsing.html#named-character-reference-state
 */
func isEntityInAttributeInvalidEnd(code byte) bool {
	return code == EQUALS || isAsciiAlphaNumeric(code)
}

type EntityDecoderState uint32

const (
	EntityDecoderStateEntityStart EntityDecoderState = iota
	EntityDecoderStateNumericStart
	EntityDecoderStateNumericDecimal
	EntityDecoderStateNumericHex
	EntityDecoderStateNamedEntity
)

type DecodingMode uint32

const (
	/** Entities in text nodes that can end with any character. */
	DecodingModeLegacy DecodingMode = iota
	/** Only allow entities terminated with a semicolon. */
	DecodingModeStrict
	/** Entities in attributes have limitations on ending characters. */
	DecodingModeAttribute
)

/**
 * Producers for character reference errors as defined in the HTML spec.
 */
type EntityErrorProducer interface {
	missingSemicolonAfterCharacterReference()
	absenceOfDigitsInNumericCharacterReference(consumedCharacters int)
	validateNumericCharacterReference(code rune)
}

type EntityDecoder struct {
	state EntityDecoderState
	/** Characters that were consumed while parsing an entity. */
	consumed int
	/**
	 * The result of the entity.
	 *
	 * Either the result index of a numeric entity, or the codepoint of a
	 * numeric entity.
	 */
	result rune

	/** The current index in the decode tree. */
	treeIndex int
	/** The number of characters that were consumed in excess. */
	excess int
	/** The mode in which the decoder is operating. */
	decodeMode DecodingMode

	decodeTree []uint16

	emitCodePoint func(cp rune, consumed int)
	errors        EntityErrorProducer
}

func NewEntityDecoder(decodeTree []uint16, emitCodePoint func(cp rune, consumed int), errors EntityErrorProducer) *EntityDecoder {
	return &EntityDecoder{
		state:         EntityDecoderStateEntityStart,
		consumed:      1,
		result:        0,
		treeIndex:     0,
		excess:        1,
		decodeMode:    DecodingModeStrict,
		decodeTree:    decodeTree,
		emitCodePoint: emitCodePoint,
		errors:        errors,
	}
}

/** Resets the instance to make it reusable. */
func (d *EntityDecoder) StartEntity(decodeMode DecodingMode) {
	d.decodeMode = decodeMode
	d.state = EntityDecoderStateEntityStart
	d.result = 0
	d.treeIndex = 0
	d.excess = 1
	d.consumed = 1
}

/**
* Write an entity to the decoder. This can be called multiple times with partial entities.
* If the entity is incomplete, the decoder will return -1.
*
* Mirrors the implementation of `getDecoder`, but with the ability to stop decoding if the
* entity is incomplete, and resume when the next string is written.
*
* @param input The string containing the entity (or a continuation of the entity).
* @param offset The offset at which the entity begins. Should be 0 if this is not the first call.
* @returns The number of characters that were consumed, or -1 if the entity is incomplete.
 */
func (d *EntityDecoder) Write(input []byte, offset int) int {
	switch d.state {
	case EntityDecoderStateEntityStart:
		{
			if input[offset] == NUM {
				d.state = EntityDecoderStateNumericStart
				d.consumed += 1
				return d.stateNumericStart(input, offset+1)
			}
			d.state = EntityDecoderStateNamedEntity
			return d.stateNamedEntity(input, offset)
		}

	case EntityDecoderStateNumericStart:
		{
			return d.stateNumericStart(input, offset)
		}

	case EntityDecoderStateNumericDecimal:
		{
			return d.stateNumericDecimal(input, offset)
		}

	case EntityDecoderStateNumericHex:
		{
			return d.stateNumericHex(input, offset)
		}

	case EntityDecoderStateNamedEntity:
		{
			return d.stateNamedEntity(input, offset)
		}
	}

	return -1
}

/**
* Switches between the numeric decimal and hexadecimal states.
*
* Equivalent to the `Numeric character reference state` in the HTML spec.
*
* @param input The string containing the entity (or a continuation of the entity).
* @param offset The current offset.
* @returns The number of characters that were consumed, or -1 if the entity is incomplete.
 */
func (d *EntityDecoder) stateNumericStart(input []byte, offset int) int {
	if offset >= len(input) {
		return -1
	}

	if (input[offset] | TO_LOWER_BIT) == LOWER_X {
		d.state = EntityDecoderStateNumericHex
		d.consumed += 1
		return d.stateNumericHex(input, offset+1)
	}

	d.state = EntityDecoderStateNumericDecimal
	return d.stateNumericDecimal(input, offset)
}

/**
* Parses a hexadecimal numeric entity.
*
* Equivalent to the `Hexademical character reference state` in the HTML spec.
*
* @param input The string containing the entity (or a continuation of the entity).
* @param offset The current offset.
* @returns The number of characters that were consumed, or -1 if the entity is incomplete.
 */
func (d *EntityDecoder) stateNumericHex(input []byte, offset int) int {
	for offset < len(input) {
		char := input[offset]
		if isNumber(char) || isHexadecimalCharacter(char) {
			// Convert hex digit to value (0-15); 'a'/'A' -> 10.
			var digit byte
			if char < NINE {
				digit = char - ZERO
			} else {
				digit = (char | TO_LOWER_BIT) - LOWER_A + 10
			}

			d.result = d.result*16 + rune(digit)
			d.consumed++
			offset++
		} else {
			return d.emitNumericEntity(char, 3)
		}
	}
	return -1 // Incomplete entity
}

/**
* Parses a decimal numeric entity.
*
* Equivalent to the `Decimal character reference state` in the HTML spec.
*
* @param input The string containing the entity (or a continuation of the entity).
* @param offset The current offset.
* @returns The number of characters that were consumed, or -1 if the entity is incomplete.
 */
func (d *EntityDecoder) stateNumericDecimal(input []byte, offset int) int {
	for offset < len(input) {
		char := input[offset]
		if isNumber(char) {
			d.result = d.result*10 + rune(char-ZERO)
			d.consumed++
			offset++
		} else {
			return d.emitNumericEntity(char, 2)
		}
	}
	return -1 // Incomplete entity
}

/**
* Validate and emit a numeric entity.
*
* Implements the logic from the `Hexademical character reference start
* state` and `Numeric character reference end state` in the HTML spec.
*
* @param lastCp The last code point of the entity. Used to see if the
*               entity was terminated with a semicolon.
* @param expectedLength The minimum number of characters that should be
*                       consumed. Used to validate that at least one digit
*                       was consumed.
* @returns The number of characters that were consumed.
 */
func (d *EntityDecoder) emitNumericEntity(lastCp byte, expectedLength int) int {
	// Ensure we consumed at least one digit.
	if d.consumed <= expectedLength {
		if d.errors != nil {
			d.errors.absenceOfDigitsInNumericCharacterReference(d.consumed)
		}
		return 0
	}

	// Figure out if this is a legit end of the entity
	if lastCp == SEMI {
		d.consumed += 1
	} else if d.decodeMode == DecodingModeStrict {
		return 0
	}

	d.emitCodePoint(replaceCodePoint(d.result), d.consumed)

	if d.errors != nil {
		if lastCp != SEMI {
			d.errors.missingSemicolonAfterCharacterReference()
		}

		d.errors.validateNumericCharacterReference(d.result)
	}

	return d.consumed
}

/**
* Parses a named entity.
*
* Equivalent to the `Named character reference state` in the HTML spec.
*
* @param input The string containing the entity (or a continuation of the entity).
* @param offset The current offset.
* @returns The number of characters that were consumed, or -1 if the entity is incomplete.
 */
func (d *EntityDecoder) stateNamedEntity(input []byte, offset int) int {
	current := d.decodeTree[d.treeIndex]
	// The mask is the number of bytes of the value, including the current byte.
	valueLength := (current & BIN_TRIE_FLAGS_VALUE_LENGTH) >> 14

	for offset < len(input) {
		char := input[offset]

		d.treeIndex = determineBranch(
			d.decodeTree,
			current,
			d.treeIndex+int(max(1, valueLength)),
			char,
		)

		if d.treeIndex < 0 {
			if d.result == 0 ||
				// If we are parsing an attribute
				(d.decodeMode == DecodingModeAttribute &&
					// We shouldn't have consumed any characters after the entity,
					(valueLength == 0 ||
						// And there should be no invalid characters.
						isEntityInAttributeInvalidEnd(char))) {
				return 0
			} else {
				return d.emitNotTerminatedNamedEntity()
			}
		}

		current = d.decodeTree[d.treeIndex]
		valueLength = (current & BIN_TRIE_FLAGS_VALUE_LENGTH) >> 14

		// If the branch is a value, store it and continue
		if valueLength != 0 {
			// If the entity is terminated by a semicolon, we are done.
			if char == SEMI {
				return d.emitNamedEntityData(
					rune(d.treeIndex),
					valueLength,
					d.consumed+d.excess,
				)
			}

			// If we encounter a non-terminated (legacy) entity while parsing strictly, then ignore it.
			if d.decodeMode != DecodingModeStrict {
				d.result = rune(d.treeIndex)
				d.consumed += d.excess
				d.excess = 0
			}
		}
		// Increment offset & excess for next iteration
		offset++
		d.excess++
	}

	return -1
}

/**
* Emit a named entity that was not terminated with a semicolon.
*
* @returns The number of characters consumed.
 */
func (d *EntityDecoder) emitNotTerminatedNamedEntity() int {
	valueLength := (d.decodeTree[d.result] & BIN_TRIE_FLAGS_VALUE_LENGTH) >> 14

	d.emitNamedEntityData(d.result, valueLength, d.consumed)
	if d.errors != nil {
		d.errors.missingSemicolonAfterCharacterReference()
	}

	return d.consumed
}

/**
* Emit a named entity.
*
* @param result The index of the entity in the decode tree.
* @param valueLength The number of bytes in the entity.
* @param consumed The number of characters consumed.
*
* @returns The number of characters consumed.
 */
func (d *EntityDecoder) emitNamedEntityData(
	result rune,
	valueLength uint16,
	consumed int,
) int {
	if valueLength == 1 {
		d.emitCodePoint(rune(d.decodeTree[result] & ^uint16(BIN_TRIE_FLAGS_VALUE_LENGTH)), consumed)
	} else {
		d.emitCodePoint(rune(d.decodeTree[result+1]), consumed)
	}
	if valueLength == 3 {
		// For multi-byte values, we need to emit the second byte.
		d.emitCodePoint(rune(d.decodeTree[result+2]), consumed)
	}

	return consumed
}

/**
* Signal to the parser that the end of the input was reached.
*
* Remaining data will be emitted and relevant errors will be produced.
*
* @returns The number of characters consumed.
 */
func (d *EntityDecoder) End() int {
	switch d.state {
	case EntityDecoderStateNamedEntity:
		{
			// Emit a named entity if we have one.
			if d.result != 0 && (d.decodeMode != DecodingModeAttribute || d.result == rune(d.treeIndex)) {
				return d.emitNotTerminatedNamedEntity()
			} else {
				return 0
			}
		}
	// Otherwise, emit a numeric entity if we have one.
	case EntityDecoderStateNumericDecimal:
		{
			return d.emitNumericEntity(0, 2)
		}
	case EntityDecoderStateNumericHex:
		{
			return d.emitNumericEntity(0, 3)
		}
	case EntityDecoderStateNumericStart:
		{
			if d.errors != nil {
				d.errors.absenceOfDigitsInNumericCharacterReference(d.consumed)
			}
			return 0
		}
	case EntityDecoderStateEntityStart:
		{
			// Return 0 if we have no entity.
			return 0
		}
	}

	return -1
}

/**
 * Creates a function that decodes entities in a string.
 *
 * @param decodeTree The decode tree.
 * @returns A function that decodes entities in a string.
 */
func getDecoder(decodeTree []uint16) func([]byte, DecodingMode) []byte {
	returnValue := make([]byte, 0, 100)
	decoder := NewEntityDecoder(
		decodeTree,
		func(data rune, consumed int) { returnValue = utf8.AppendRune(returnValue, replaceCodePoint(data)) },
		nil,
	)

	return func(
		input []byte,
		decodeMode DecodingMode,
	) []byte {
		lastIndex := 0
		offset := 0

		for {
			offset = indexOf(input, byte('&'), offset)
			if offset < 0 {
				break
			}

			returnValue = append(returnValue, input[lastIndex:offset]...)

			decoder.StartEntity(decodeMode)

			length := decoder.Write(
				input,
				// Skip the "&"
				offset+1,
			)

			if length < 0 {
				lastIndex = offset + decoder.End()
				break
			}

			lastIndex = offset + length
			// If `length` is 0, skip the current `&` and continue.
			if length == 0 {
				offset = lastIndex + 1
			} else {
				offset = lastIndex
			}
		}

		result := append(returnValue, input[lastIndex:]...)

		// Make sure we don't keep a reference to the final string.
		returnValue = make([]byte, 0, 100)

		return result
	}
}

func indexOf(data []byte, b byte, offset int) int {
	if offset < 0 {
		offset = 0
	}

	for i := offset; i < len(data); i++ {
		if data[i] == b {
			return i
		}
	}

	return -1
}

/**
 * Determines the branch of the current node that is taken given the current
 * character. This function is used to traverse the trie.
 *
 * @param decodeTree The trie.
 * @param current The current node.
 * @param nodeIdx The index right after the current node and its value.
 * @param char The current character.
 * @returns The index of the next node, or -1 if no branch is taken.
 */
func determineBranch(
	decodeTree []uint16,
	current uint16,
	nodeIndex int,
	char byte,
) int {
	branchCount := (current & BIN_TRIE_FLAGS_BRANCH_LENGTH) >> 7
	jumpOffset := current & BIN_TRIE_FLAGS_JUMP_TABLE

	// Case 1: Single branch encoded in jump offset
	if branchCount == 0 {
		if jumpOffset != 0 && uint16(char) == jumpOffset {
			return nodeIndex
		} else {
			return -1
		}
	}

	// Case 2: Multiple branches encoded in jump table
	if jumpOffset > 0 {
		value := int(char) - int(jumpOffset)

		if value < 0 || value >= int(branchCount) {
			return -1
		} else {
			return int(decodeTree[nodeIndex+value]) - 1
		}
	}

	// Case 3: Multiple branches encoded in dictionary

	// Binary search for the character.
	var lo int = nodeIndex
	var hi int = lo + int(branchCount) - 1

	for lo <= hi {
		mid := uint16(lo+hi) >> 1
		midValue := decodeTree[mid]

		if midValue < uint16(char) {
			lo = int(mid) + 1
		} else if midValue > uint16(char) {
			hi = int(mid) - 1
		} else {
			return int(decodeTree[int(mid)+int(branchCount)])
		}
	}

	return -1
}

var htmlDecoder = getDecoder(HtmlDecodeTree)
var xmlDecoder = getDecoder(XmlDecodeTree)

/**
 * Decodes an HTML string.
 *
 * @param htmlString The string to decode.
 * @param mode The decoding mode.
 * @returns The decoded string.
 */
func DecodeHTMLLegacy(htmlString string) string {
	return string(htmlDecoder([]byte(htmlString), DecodingModeLegacy))
}

/**
 * Decodes an HTML string in an attribute.
 *
 * @param htmlAttribute The string to decode.
 * @returns The decoded string.
 */
func DecodeHTMLAttribute(htmlAttribute string) string {
	return string(htmlDecoder([]byte(htmlAttribute), DecodingModeAttribute))
}

/**
 * Decodes an HTML string, requiring all entities to be terminated by a semicolon.
 *
 * @param htmlString The string to decode.
 * @returns The decoded string.
 */
func DecodeHTMLStrict(htmlString string) string {
	return string(htmlDecoder([]byte(htmlString), DecodingModeStrict))
}

/**
 * Decodes an XML string, requiring all entities to be terminated by a semicolon.
 *
 * @param xmlString The string to decode.
 * @returns The decoded string.
 */
func DecodeXML(xmlString string) string {
	return string(xmlDecoder([]byte(xmlString), DecodingModeStrict))
}
