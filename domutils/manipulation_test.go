package domutils

import (
	"testing"

	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/serializer"
	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T {
	return &v
}

func stringify(node *dom.Node) string {
	return serializer.Render([]*dom.Node{node}, &serializer.DomSerializerOptions{XmlMode: ptr("true"), SelfClosingTags: ptr(true)})
}

func TestManipulation(t *testing.T) {
	t.Run("Append: should not be duplicated when called twice", func(t *testing.T) {
		dom := parseDOM("<div><p><img/></p><p><object/></p></div>")[0]
		child := parseDOM("<span></span>")[0]
		parents := dom.Children

		Append(parents[0].Children[0], child)

		assert.Equal(t, `<div><p><img/><span/></p><p><object/></p></div>`, stringify(dom))

		Append(parents[1].Children[0], child)

		assert.Len(t, parents[0].Children, 1)
		assert.Len(t, parents[1].Children, 2)

		assert.Equal(t, `<div><p><img/></p><p><object/><span/></p></div>`, stringify(dom))
	})

	t.Run("AppendChild: should not be duplicated when called twice", func(t *testing.T) {
		dom := parseDOM("<div><p><img/></p><p><object/></p></div>")[0]
		child := parseDOM("<span></span>")[0]
		parents := dom.Children

		AppendChild(parents[0], child)

		assert.Equal(t, `<div><p><img/><span/></p><p><object/></p></div>`, stringify(dom))

		AppendChild(parents[1], child)

		assert.Len(t, parents[0].Children, 1)
		assert.Len(t, parents[1].Children, 2)

		assert.Equal(t, `<div><p><img/></p><p><object/><span/></p></div>`, stringify(dom))
	})

	t.Run("RemoveElement: should correctly remove element", func(t *testing.T) {
		dom := parseDOM("<div><p><img/><object/></p><p></p></div>")[0]
		parents := dom.Children
		image := parents[0].Children[0]

		RemoveElement(image)

		assert.Nil(t, image.NextSibling)
		assert.Nil(t, image.PreviousSibling)
		assert.Nil(t, image.Parent)

		assert.Equal(t, `<div><p><object/></p><p/></div>`, stringify(dom))

		AppendChild(parents[1], image)

		assert.Len(t, parents[0].Children, 1)
		assert.Len(t, parents[1].Children, 1)

		assert.Equal(t, `<div><p><object/></p><p><img/></p></div>`, stringify(dom))
	})

	t.Run("should not be duplicated when called twice", func(t *testing.T) {
		dom := parseDOM("<div><p><img/></p><p><object/></p></div>")[0]
		child := parseDOM("<span></span>")[0]
		parents := dom.Children
		Prepend(parents[0].Children[0], child)

		assert.Equal(t, `<div><p><span/><img/></p><p><object/></p></div>`, stringify(dom))

		Prepend(parents[1].Children[0], child)
		assert.Len(t, parents[0].Children, 1)
		assert.Len(t, parents[1].Children, 2)

		assert.Equal(t, `<div><p><img/></p><p><span/><object/></p></div>`, stringify(dom))
	})

	t.Run("PrependChild: should not be duplicated when called twice", func(t *testing.T) {
		dom := parseDOM("<div><p><img/></p><p><object/></p></div>")[0]
		child := parseDOM("<span></span>")[0]
		parents := dom.Children

		PrependChild(parents[0], child)

		assert.Equal(t, `<div><p><span/><img/></p><p><object/></p></div>`, stringify(dom))

		PrependChild(parents[1], child)

		assert.Len(t, parents[0].Children, 1)
		assert.Len(t, parents[1].Children, 2)

		assert.Equal(t, `<div><p><img/></p><p><span/><object/></p></div>`, stringify(dom))
	})

	t.Run("ReplaceElement: should allow replaced elements to be appended later (#966)", func(t *testing.T) {
		div := parseDOM("<div><p>")[0]
		template := parseDOM("<template></template>")[0]
		p := div.Children[0]

		// We want to wrap the inner <p> in a <template>
		ReplaceElement(p, template)
		AppendChild(template, p)

		assert.Equal(t, `<div><template><p/></template></div>`, stringify(div))
	})
}
