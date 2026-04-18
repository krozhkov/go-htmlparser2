package domutils

import (
	"testing"

	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/parser"
	"github.com/stretchr/testify/assert"
)

func parseDOM(str string) []*dom.Node {
	return dom.ParseDOM(str, &parser.ParserOptions{LowerCaseAttributeNames: true, DecodeEntities: true, RecognizeSelfClosing: true})
}

func TestTraversal(t *testing.T) {
	t.Run("getSiblings", func(t *testing.T) {
		t.Run("returns an element's siblings", func(t *testing.T) {
			dom := parseDOM("<div><h1></h1><p><p><p></div>")[0]

			assert.Len(t, GetSiblings(dom.Children[1]), 4)
		})

		t.Run("returns a root element's siblings", func(t *testing.T) {
			dom := parseDOM("<h1></h1><p><p><p>")

			for _, node := range dom {
				node.Parent = nil
			}

			assert.Len(t, GetSiblings(dom[2]), 4)
		})
	})

	t.Run("hasAttrib", func(t *testing.T) {
		t.Run("doesn't throw on text nodes", func(t *testing.T) {
			assert.False(t, HasAttrib(parseDOM("textnode")[0], "some-attrib"))
		})

		t.Run("returns `false` for Object prototype properties", func(t *testing.T) {
			assert.False(t, HasAttrib(parseDOM("<div><h1></h1>test<p></p></div>")[0], "constructor"))
		})

		t.Run("should return `false` for \"null\" values", func(t *testing.T) {
			div := parseDOM("<div class=test>")[0]

			assert.True(t, HasAttrib(div, "class"))

			div.Attribs.Delete("class")

			assert.False(t, HasAttrib(div, "class"))
		})
	})

	t.Run("nextElementSibling", func(t *testing.T) {
		t.Run("return Element if found", func(t *testing.T) {
			dom := parseDOM("<div><h1></h1>test<p></p></div>")[0]
			firstNode := dom.FirstChild()

			next := NextElementSibling(firstNode)
			assert.Equal(t, "p", next.Name)
		})
		t.Run("return null if not found", func(t *testing.T) {
			dom := parseDOM("<div><p></p>test</div>")[0]
			firstNode := dom.FirstChild()

			next := NextElementSibling(firstNode)
			assert.Nil(t, next)
		})
		t.Run("does not ignore script tags", func(t *testing.T) {
			dom := parseDOM("<div><p></p><script></script><p></div>")[0]
			firstNode := dom.FirstChild()

			next := NextElementSibling(firstNode)
			assert.Equal(t, "script", next.Name)
		})
	})

	t.Run("prevElementSibling", func(t *testing.T) {
		t.Run("return Element if found", func(t *testing.T) {
			dom := parseDOM("<div><h1></h1>test<p></p></div>")[0]
			lastNode := dom.Children[2]

			prev := PrevElementSibling(lastNode)
			assert.Equal(t, "h1", prev.Name)
		})
		t.Run("return null if not found", func(t *testing.T) {
			dom := parseDOM("<div>test<p></p></div>")[0]
			lastNode := dom.Children[1]

			assert.Nil(t, PrevElementSibling(lastNode))
		})
		t.Run("does not ignore script tags", func(t *testing.T) {
			dom := parseDOM("<div><p></p><script></script><p></p></div>")[0]
			lastNode := dom.Children[2]

			prev := PrevElementSibling(lastNode)
			assert.Equal(t, "script", prev.Name)
		})
	})

	t.Run("getAttributeValue", func(t *testing.T) {
		t.Run("returns the attribute value", func(t *testing.T) {
			dom := parseDOM("<div class='test'>")[0]
			assert.Equal(t, "test", GetAttributeValue(dom, "class"))
		})
		t.Run("returns undefined if attribute does not exist", func(t *testing.T) {
			assert.Equal(t, "", GetAttributeValue(parseDOM("<div>")[0], "id"))
		})
		t.Run("should return undefined if a random node is passed", func(t *testing.T) {
			assert.Equal(t, "", GetAttributeValue(parseDOM("TEXT")[0], "id"))
		})
	})

	t.Run("getName", func(t *testing.T) {
		t.Run("returns the name of the element", func(t *testing.T) {
			assert.Equal(t, "div", GetName(parseDOM("<div>")[0]))
		})
		t.Run("should return undefined if a random node is passed", func(t *testing.T) {
			assert.Equal(t, "", GetName(parseDOM("TEXT")[0]))
		})
	})
}
