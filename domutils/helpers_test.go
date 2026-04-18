package domutils

import (
	"testing"

	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/parser"
	"github.com/stretchr/testify/assert"
)

func TestHelpers(t *testing.T) {
	t.Run("removeSubsets", func(t *testing.T) {
		parsed := dom.ParseDocument("<div><p><span></span></p><p></p></div>", &parser.ParserOptions{LowerCaseAttributeNames: true, DecodeEntities: true, RecognizeSelfClosing: true})
		node := parsed.Children[0]

		t.Run("removes identical trees", func(t *testing.T) {
			assert.Len(t, RemoveSubsets([]*dom.Node{node, node}), 1)
		})

		t.Run("Removes subsets found first", func(t *testing.T) {
			firstChild := node.Children[0]
			matches := RemoveSubsets([]*dom.Node{node, firstChild.Children[0]})
			assert.Len(t, matches, 1)
		})

		t.Run("Removes subsets found last", func(t *testing.T) {
			assert.Len(t, RemoveSubsets([]*dom.Node{node.Children[0], node}), 1)
		})

		t.Run("Does not remove unique trees", func(t *testing.T) {
			assert.Len(t, RemoveSubsets([]*dom.Node{node.Children[0], node.Children[1]}), 2)
		})
	})
}
