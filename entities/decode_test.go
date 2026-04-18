package entities

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type decodeTest struct {
	input  string
	output string
}

func TestDecode(t *testing.T) {
	var testCases = []decodeTest{
		{input: "&amp;amp;", output: "&amp;"},
		{input: "&amp;#38;", output: "&#38;"},
		{input: "&amp;#x26;", output: "&#x26;"},
		{input: "&amp;#X26;", output: "&#X26;"},
		{input: "&#38;#38;", output: "&#38;"},
		{input: "&#x26;#38;", output: "&#38;"},
		{input: "&#X26;#38;", output: "&#38;"},
		{input: "&#x3a;", output: ":"},
		{input: "&#x3A;", output: ":"},
		{input: "&#X3a;", output: ":"},
		{input: "&#X3A;", output: ":"},
		{input: "&#", output: "&#"},
		{input: "&>", output: "&>"},
		{input: "id=770&#anchor", output: "id=770&#anchor"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("should XML decode %q", tc.input), func(t *testing.T) {
			assert.Equal(t, tc.output, DecodeXML(tc.input))
		})

		t.Run(fmt.Sprintf("should HTML decode %q", tc.input), func(t *testing.T) {
			assert.Equal(t, tc.output, DecodeHTMLLegacy(tc.input))
		})
	}

	t.Run("should HTML decode partial legacy entity", func(t *testing.T) {
		assert.Equal(t, "&timesbar", DecodeHTMLStrict("&timesbar"))
		assert.Equal(t, "×bar", DecodeHTMLLegacy("&timesbar"))
	})

	t.Run("should HTML decode legacy entities according to spec", func(t *testing.T) {
		assert.Equal(t, "?&image_uri=1&ℑ=2&image=3", DecodeHTMLLegacy("?&image_uri=1&ℑ=2&image=3"))
	})

	t.Run("should back out of legacy entities", func(t *testing.T) {
		assert.Equal(t, "&a", DecodeHTMLLegacy("&ampa"))
	})

	t.Run("should not parse numeric entities in strict mode", func(t *testing.T) {
		assert.Equal(t, "&#55", DecodeHTMLStrict("&#55"))
	})

	t.Run("should parse &nbsp followed by < (#852)", func(t *testing.T) {
		assert.Equal(t, "\u00A0<", DecodeHTMLLegacy("&nbsp<"))
	})

	t.Run("should decode trailing legacy entities", func(t *testing.T) {
		assert.Equal(t, "⨱×bar", DecodeHTMLLegacy("&timesbar;&timesbar"))
	})

	t.Run("should decode multi-byte entities", func(t *testing.T) {
		assert.Equal(t, "≧̸", DecodeHTMLLegacy("&NotGreaterFullEqual;"))
	})

	t.Run("should not decode legacy entities followed by text in attribute mode", func(t *testing.T) {
		assert.Equal(t, "¬", DecodeHTMLAttribute("&not"))

		assert.Equal(t, "&noti", DecodeHTMLAttribute("&noti"))

		assert.Equal(t, "&not=", DecodeHTMLAttribute("&not="))

		assert.Equal(t, "&notp", DecodeHTMLAttribute("&notp"))
		assert.Equal(t, "&notP", DecodeHTMLAttribute("&notP"))
		assert.Equal(t, "&not3", DecodeHTMLAttribute("&not3"))
	})
}

type CallbacksMock struct {
	mock.Mock
}

func (m *CallbacksMock) emitCodePoint(data rune, consumed int) {
	m.Called(data, consumed)
}

func NewCallbacksMock() *CallbacksMock {
	m := new(CallbacksMock)
	m.On("emitCodePoint", mock.Anything, mock.Anything).Return()
	return m
}

type EntityErrorProducerMock struct {
	mock.Mock
}

func (m *EntityErrorProducerMock) missingSemicolonAfterCharacterReference() {
	m.Called()
}
func (m *EntityErrorProducerMock) absenceOfDigitsInNumericCharacterReference(consumedCharacters int) {
	m.Called(consumedCharacters)
}
func (m *EntityErrorProducerMock) validateNumericCharacterReference(code rune) {
	m.Called(code)
}

func NewEntityErrorProducerMock() *EntityErrorProducerMock {
	m := new(EntityErrorProducerMock)
	m.On("missingSemicolonAfterCharacterReference").Return()
	m.On("absenceOfDigitsInNumericCharacterReference", mock.Anything).Return()
	m.On("validateNumericCharacterReference", mock.Anything).Return()
	return m
}

func TestEntityDecoder(t *testing.T) {
	t.Run("should decode decimal entities", func(t *testing.T) {
		callback := NewCallbacksMock()
		decoder := NewEntityDecoder(
			HtmlDecodeTree,
			callback.emitCodePoint,
			nil,
		)

		assert.Equal(t, -1, decoder.Write([]byte("&#5"), 1))
		assert.Equal(t, 5, decoder.Write([]byte("8;"), 0))
		callback.AssertNumberOfCalls(t, "emitCodePoint", 1)
		callback.AssertCalled(t, "emitCodePoint", ':', 5)
	})

	t.Run("should decode hex entities", func(t *testing.T) {
		callback := NewCallbacksMock()
		decoder := NewEntityDecoder(
			HtmlDecodeTree,
			callback.emitCodePoint,
			nil,
		)

		assert.Equal(t, 6, decoder.Write([]byte("&#x3a;"), 1))
		callback.AssertNumberOfCalls(t, "emitCodePoint", 1)
		callback.AssertCalled(t, "emitCodePoint", ':', 6)
	})

	t.Run("should decode named entities", func(t *testing.T) {
		callback := NewCallbacksMock()
		decoder := NewEntityDecoder(
			HtmlDecodeTree,
			callback.emitCodePoint,
			nil,
		)

		assert.Equal(t, 5, decoder.Write([]byte("&amp;"), 1))
		callback.AssertNumberOfCalls(t, "emitCodePoint", 1)
		callback.AssertCalled(t, "emitCodePoint", '&', 5)
	})

	t.Run("should decode legacy entities", func(t *testing.T) {
		callback := NewCallbacksMock()
		decoder := NewEntityDecoder(
			HtmlDecodeTree,
			callback.emitCodePoint,
			nil,
		)

		decoder.StartEntity(DecodingModeLegacy)

		assert.Equal(t, -1, decoder.Write([]byte("&amp"), 1))

		callback.AssertNotCalled(t, "emitCodePoint")

		assert.Equal(t, 4, decoder.End())
		callback.AssertNumberOfCalls(t, "emitCodePoint", 1)
		callback.AssertCalled(t, "emitCodePoint", '&', 4)
	})

	t.Run("should decode named entity written character by character", func(t *testing.T) {
		callback := NewCallbacksMock()
		decoder := NewEntityDecoder(
			HtmlDecodeTree,
			callback.emitCodePoint,
			nil,
		)

		for _, c := range "amp" {
			assert.Equal(t, -1, decoder.Write([]byte{byte(c)}, 0))
		}

		assert.Equal(t, 5, decoder.Write([]byte(";"), 0))
		callback.AssertNumberOfCalls(t, "emitCodePoint", 1)
		callback.AssertCalled(t, "emitCodePoint", '&', 5)
	})

	t.Run("should decode numeric entity written character by character", func(t *testing.T) {
		callback := NewCallbacksMock()
		decoder := NewEntityDecoder(
			HtmlDecodeTree,
			callback.emitCodePoint,
			nil,
		)

		for _, c := range "#x3a" {
			assert.Equal(t, -1, decoder.Write([]byte{byte(c)}, 0))
		}

		assert.Equal(t, 6, decoder.Write([]byte(";"), 0))
		callback.AssertNumberOfCalls(t, "emitCodePoint", 1)
		callback.AssertCalled(t, "emitCodePoint", ':', 6)
	})

	t.Run("should decode hex entities across several chunks", func(t *testing.T) {
		callback := NewCallbacksMock()
		decoder := NewEntityDecoder(
			HtmlDecodeTree,
			callback.emitCodePoint,
			nil,
		)

		for _, chunk := range []string{"#x", "cf", "ff", "d"} {
			assert.Equal(t, -1, decoder.Write([]byte(chunk), 0))
		}

		assert.Equal(t, 9, decoder.Write([]byte(";"), 0))
		callback.AssertNumberOfCalls(t, "emitCodePoint", 1)
		callback.AssertCalled(t, "emitCodePoint", rune(0xc_ff_fd), 9)
	})

	t.Run("should not fail if nothing is written", func(t *testing.T) {
		callback := NewCallbacksMock()

		decoder := NewEntityDecoder(
			HtmlDecodeTree,
			callback.emitCodePoint,
			nil,
		)

		assert.Equal(t, 0, decoder.End())

		callback.AssertNotCalled(t, "emitCodePoint")
	})

	/*
	 * Focused tests exercising early exit paths inside a compact run in the real trie.
	 * Discovered prefix: "zi" followed by compact run "grarr"; mismatching inside this run should
	 * return 0 with no emission (result still 0).
	 */
	t.Run("compact run mismatches", func(t *testing.T) {
		t.Run("first run character mismatch returns 0", func(t *testing.T) {
			callback := NewCallbacksMock()
			d := NewEntityDecoder(
				HtmlDecodeTree,
				callback.emitCodePoint,
				nil,
			)
			d.StartEntity(DecodingModeStrict)
			// After '&': correct prefix 'zi', wrong first run char 'X' (expected 'g').
			assert.Equal(t, 0, d.Write([]byte("ziXgrar"), 0))
			callback.AssertNotCalled(t, "emitCodePoint")
		})

		t.Run("mismatch after one correct run char returns 0", func(t *testing.T) {
			callback := NewCallbacksMock()
			d := NewEntityDecoder(
				HtmlDecodeTree,
				callback.emitCodePoint,
				nil,
			)
			d.StartEntity(DecodingModeStrict)
			// 'zig' matches prefix + first run char; next char 'X' mismatches expected 'r'.
			assert.Equal(t, 0, d.Write([]byte("zigXarr"), 0))
			callback.AssertNotCalled(t, "emitCodePoint")
		})

		t.Run("mismatch after two correct run chars returns 0", func(t *testing.T) {
			callback := NewCallbacksMock()
			d := NewEntityDecoder(
				HtmlDecodeTree,
				callback.emitCodePoint,
				nil,
			)
			d.StartEntity(DecodingModeStrict)
			// 'zigr' matches prefix + first two run chars; next char 'X' mismatches expected 'a'.
			assert.Equal(t, 0, d.Write([]byte("zigrXrr"), 0))
			callback.AssertNotCalled(t, "emitCodePoint")
		})
	})

	t.Run("errors", func(t *testing.T) {
		t.Run("should produce an error for a named entity without a semicolon", func(t *testing.T) {
			errorHandlers := NewEntityErrorProducerMock()
			callback := NewCallbacksMock()
			decoder := NewEntityDecoder(
				HtmlDecodeTree,
				callback.emitCodePoint,
				errorHandlers,
			)

			decoder.StartEntity(DecodingModeLegacy)
			assert.Equal(t, 5, decoder.Write([]byte("&amp;"), 1))
			callback.AssertNumberOfCalls(t, "emitCodePoint", 1)
			callback.AssertCalled(t, "emitCodePoint", '&', 5)
			errorHandlers.AssertNotCalled(t, "missingSemicolonAfterCharacterReference")

			decoder.StartEntity(DecodingModeLegacy)
			assert.Equal(t, -1, decoder.Write([]byte("&amp"), 1))
			assert.Equal(t, 4, decoder.End())

			callback.AssertNumberOfCalls(t, "emitCodePoint", 2)
			callback.AssertCalled(t, "emitCodePoint", '&', 4)
			errorHandlers.AssertNumberOfCalls(t, "missingSemicolonAfterCharacterReference", 1)
		})

		t.Run("should produce an error for a numeric entity without a semicolon", func(t *testing.T) {
			errorHandlers := NewEntityErrorProducerMock()
			callback := NewCallbacksMock()
			decoder := NewEntityDecoder(
				HtmlDecodeTree,
				callback.emitCodePoint,
				errorHandlers,
			)

			decoder.StartEntity(DecodingModeLegacy)
			assert.Equal(t, -1, decoder.Write([]byte("&#x3a"), 1))
			assert.Equal(t, 5, decoder.End())

			callback.AssertNumberOfCalls(t, "emitCodePoint", 1)
			callback.AssertCalled(t, "emitCodePoint", rune(0x3a), 5)
			errorHandlers.AssertNumberOfCalls(t, "missingSemicolonAfterCharacterReference", 1)
			errorHandlers.AssertNotCalled(t, "absenceOfDigitsInNumericCharacterReference")
			errorHandlers.AssertNumberOfCalls(t, "validateNumericCharacterReference", 1)
			errorHandlers.AssertCalled(t, "validateNumericCharacterReference", rune(0x3a))
		})

		t.Run("should produce an error for numeric entities without digits", func(t *testing.T) {
			errorHandlers := NewEntityErrorProducerMock()
			callback := NewCallbacksMock()
			decoder := NewEntityDecoder(
				HtmlDecodeTree,
				callback.emitCodePoint,
				errorHandlers,
			)

			decoder.StartEntity(DecodingModeLegacy)
			assert.Equal(t, -1, decoder.Write([]byte("&#"), 1))
			assert.Equal(t, 0, decoder.End())

			callback.AssertNotCalled(t, "emitCodePoint")
			errorHandlers.AssertNotCalled(t, "missingSemicolonAfterCharacterReference")
			errorHandlers.AssertNumberOfCalls(t, "absenceOfDigitsInNumericCharacterReference", 1)
			errorHandlers.AssertCalled(t, "absenceOfDigitsInNumericCharacterReference", 2)
			errorHandlers.AssertNotCalled(t, "validateNumericCharacterReference")
		})

		t.Run("should produce an error for hex entities without digits", func(t *testing.T) {
			errorHandlers := NewEntityErrorProducerMock()
			callback := NewCallbacksMock()
			decoder := NewEntityDecoder(
				HtmlDecodeTree,
				callback.emitCodePoint,
				errorHandlers,
			)

			decoder.StartEntity(DecodingModeLegacy)
			assert.Equal(t, -1, decoder.Write([]byte("&#x"), 1))
			assert.Equal(t, 0, decoder.End())

			callback.AssertNotCalled(t, "emitCodePoint")
			errorHandlers.AssertNotCalled(t, "missingSemicolonAfterCharacterReference")
			errorHandlers.AssertNumberOfCalls(t, "absenceOfDigitsInNumericCharacterReference", 1)
			errorHandlers.AssertNotCalled(t, "validateNumericCharacterReference")
		})
	})
}
