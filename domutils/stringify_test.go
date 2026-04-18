package domutils

import (
	"strings"
	"testing"

	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/parser"
	"github.com/stretchr/testify/assert"
)

var markup = strings.Repeat("<?xml><tag1 id='asdf' class='class1'> <script>text</script> <!-- comment --> <tag2> text </tag1>", 20)

func TestStringify(t *testing.T) {
	var fixture = dom.ParseDOM(markup, &parser.ParserOptions{LowerCaseAttributeNames: true, DecodeEntities: true, RecognizeSelfClosing: true})
	t.Run("GetText correctly renders the text content", func(t *testing.T) {
		assert.Equal(t, " text   text ", GetText(fixture[1]))
	})

	t.Run("TextContent correctly renders the text content", func(t *testing.T) {
		assert.Equal(t, " text   text ", TextContent(fixture[1]))
	})

	t.Run("InnerText correctly renders the text content", func(t *testing.T) {
		assert.Equal(t, "    text ", InnerText(fixture[1]))
	})
}
