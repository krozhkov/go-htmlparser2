package entities

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type encodeTest struct {
	input string
	xml   string
	html  string
}

func TestEscape(t *testing.T) {
	var testCases = []encodeTest{
		{
			input: "asdf & ÿ ü '",
			xml:   "asdf &amp; &#xff; &#xfc; &apos;",
			html:  "asdf &amp; &yuml; &uuml; &apos;",
		},
		{
			input: "&#38;",
			xml:   "&amp;#38;",
			html:  "&amp;&num;38&semi;",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("should XML encode %q", tc.input), func(t *testing.T) {
			encodedXML := EncodeXML(tc.input)
			assert.Equal(t, tc.xml, encodedXML)
		})
	}

	t.Run("should escape HTML attribute values", func(t *testing.T) {
		escaped := EscapeAttribute("<a \" attr > & value \u00A0!")
		assert.Equal(t, "<a &quot; attr > &amp; value &nbsp;!", escaped)
	})

	t.Run("should escape HTML text", func(t *testing.T) {
		escaped := EscapeText("<a \" text > & value \u00A0!")
		assert.Equal(t, "&lt;a \" text &gt; &amp; value &nbsp;!", escaped)
	})
}
