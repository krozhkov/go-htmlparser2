package domutils

import (
	"math"
	"testing"

	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/parser"
	"github.com/stretchr/testify/assert"
)

func prepareNode(node *dom.Node) *dom.Node {
	node.Parent = nil
	node.NextSibling = nil
	node.PreviousSibling = nil

	for _, child := range node.Children {
		prepareNode(child)
	}

	return node
}

func TestLegacy(t *testing.T) {
	var doc = prepareNode(
		dom.ParseDocument(markup, &parser.ParserOptions{LowerCaseAttributeNames: true, DecodeEntities: true, RecognizeSelfClosing: true}),
	)
	var fixture = doc.Children

	// Set up expected structures
	var idAsdf = fixture[1]
	var tag2 []*dom.Node
	var typeScript []*dom.Node

	for idx := 0; idx < 20; idx++ {
		node := fixture[idx*2+1]
		tag2 = append(tag2, node.Children[5])
		typeScript = append(typeScript, node.Children[1])
	}

	t.Run("getElementById", func(t *testing.T) {
		t.Run("returns the specified node", func(t *testing.T) {
			assert.Equal(t, idAsdf, GetElementById("asdf", fixture, true))
		})
		t.Run("returns `null` for unknown IDs", func(t *testing.T) {
			assert.Nil(t, GetElementById("asdfs", fixture, true))
		})
	})

	t.Run("getElementsByClassName", func(t *testing.T) {
		t.Run("returns the specified nodes", func(t *testing.T) {
			assert.Len(t,
				GetElementsByClassName("class1", fixture, true, math.MaxInt),
				20)
		})
		t.Run("returns empty array for unknown class names", func(t *testing.T) {
			assert.Len(t,
				GetElementsByClassName("class23", fixture, true, math.MaxInt),
				0)
		})
	})

	t.Run("getElementsByTagName", func(t *testing.T) {
		t.Run("returns the specified nodes", func(t *testing.T) {
			assert.Equal(t, tag2, GetElementsByTagName("tag2", fixture, true, math.MaxInt))
		})
		t.Run("returns empty array for unknown tag names", func(t *testing.T) {
			assert.Len(t, GetElementsByTagName("tag23", fixture, true, math.MaxInt), 0)
		})
	})

	t.Run("getElementsByTagType", func(t *testing.T) {
		t.Run("returns the specified nodes", func(t *testing.T) {
			assert.Equal(t, typeScript, GetElementsByTagType(dom.ElementTypeScript, fixture, true, math.MaxInt))
		})
		t.Run("returns empty array for unknown tag types", func(t *testing.T) {
			assert.Len(t, GetElementsByTagType("video", fixture, true, math.MaxInt), 0)
		})
	})
}
