package dom

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/krozhkov/go-htmlparser2/parser"
	"github.com/stretchr/testify/assert"
)

func parse(data string, dhOptions *DomHandlerOptions) *Node {
	handler := NewDomHandler(func(err error, _ []*Node) {
		if err != nil {
			panic(err)
		}
	}, dhOptions, nil)

	var pOptions *parser.ParserOptions
	if dhOptions != nil {
		pOptions = &parser.ParserOptions{XmlMode: dhOptions.XmlMode}
	}

	parser := parser.NewParser(handler, pOptions)

	parser.End([]byte(data))

	return prepareNode(handler.Root)
}

func prepareNode(node *Node) *Node {
	if node == nil {
		return nil
	}

	node.Parent = nil
	node.NextSibling = nil
	node.PreviousSibling = nil

	if node.Attribs != nil && node.Attribs.Len() == 0 {
		node.Attribs = nil
	}

	for _, child := range node.Children {
		prepareNode(child)
	}

	return node
}

func TestNodes(t *testing.T) {
	t.Run("should serialize to a Jest snapshot", func(t *testing.T) {
		result := parse(
			"<html><!-- A Comment --><title>The Title</title><body>Hello world<input disabled type=text></body></html>",
			nil)

		snaps.MatchSnapshot(t, result)
	})

	t.Run("should be cloneable", func(t *testing.T) {
		result := parse(
			`<html><!-- A Comment -->
                <!doctype html>
                <title>The Title</title>
                <body>Hello world<input disabled type=text></body>
                <script><![CDATA[secret script]]></script>
            </html>`,
			nil)

		clone, err := result.CloneNode(true)

		assert.Nil(t, err)
		assert.Equal(t, result, prepareNode(clone))
	})

	t.Run("should not clone recursively if not asked to", func(t *testing.T) {
		result := parse("<div foo=bar><div><div>", nil)

		clone, err := result.CloneNode(true)

		assert.Nil(t, err)
		assert.Equal(t, result, prepareNode(clone))

		clone, err = result.CloneNode(false)

		assert.Nil(t, err)
		assert.NotEqual(t, result, prepareNode(clone))

		assert.Nil(t, clone.Children)
	})

	t.Run("should clone startIndex and endIndex", func(t *testing.T) {
		result := parse("<div foo=bar><div><div>", &DomHandlerOptions{WithStartIndices: true, WithEndIndices: true})

		var child *Node
		if len(result.Children) > 0 {
			child = result.Children[0]
		}

		assert.NotNil(t, child)

		clone, err := child.CloneNode(true)

		assert.Nil(t, err)
		assert.Equal(t, 0, clone.StartIndex)
		assert.Equal(t, 23, clone.EndIndex)
	})

	t.Run("should throw an error when cloning unsupported types", func(t *testing.T) {
		docType := &Node{
			Type:     ElementTypeDoctype,
			NodeType: 0,
		}

		_, err := CloneNode(docType, false)

		assert.NotNil(t, err)
		assert.Equal(t, "Not implemented yet: doctype", err.Error())
	})

	t.Run("should detect tag types", func(t *testing.T) {
		result := parse("<div foo=bar><div><div>", nil)

		var child *Node
		if len(result.Children) > 0 {
			child = result.Children[0]
		}

		assert.NotNil(t, child)

		assert.True(t, IsTag(child))
		assert.True(t, HasChildren(child))

		assert.False(t, IsCDATA(child))
		assert.False(t, IsText(child))
		assert.False(t, IsComment(child))
		assert.False(t, IsDirective(child))
		assert.False(t, IsDocument(child))
	})
}
