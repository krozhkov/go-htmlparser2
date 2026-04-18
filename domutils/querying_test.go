package domutils

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/parser"
	"github.com/stretchr/testify/assert"
)

func parseDocument(str string) *dom.Node {
	return dom.ParseDocument(str, &parser.ParserOptions{LowerCaseAttributeNames: true, DecodeEntities: true, RecognizeSelfClosing: true})
}

func TestQuerying(t *testing.T) {
	manyNodesWide := parseDocument(
		fmt.Sprintf("<body>%sText</body>", strings.Repeat("<div></div>", 200_000)),
	)
	someDeepNodes := parseDocument(
		`<body><div><div></div></div><div><p></p></div></body>`,
	)

	t.Run("find", func(t *testing.T) {
		t.Run("should accept many children without RangeError", func(t *testing.T) {
			found := Find(
				func(elem *dom.Node) bool { return elem.Type == dom.ElementTypeTag },
				[]*dom.Node{manyNodesWide},
				true,
				math.MaxInt,
			)

			assert.Len(t, found, 200_001)
		})

		t.Run("should respect limit", func(t *testing.T) {
			found := Find(
				func(elem *dom.Node) bool { return elem.Type == dom.ElementTypeTag },
				[]*dom.Node{manyNodesWide},
				true,
				20,
			)

			assert.Len(t, found, 20)
		})

		t.Run("should find text nodes", func(t *testing.T) {
			found := Find(
				func(elem *dom.Node) bool { return elem.Type == dom.ElementTypeText },
				[]*dom.Node{manyNodesWide},
				true,
				math.MaxInt,
			)

			assert.Len(t, found, 1)
		})
	})

	t.Run("findAll", func(t *testing.T) {
		t.Run("should accept many children without RangeError", func(t *testing.T) {
			assert.Len(t,
				FindAll(func(elem *dom.Node) bool { return elem.Name == "div" }, []*dom.Node{manyNodesWide}),
				200_000,
			)
		})
	})

	t.Run("filter", func(t *testing.T) {
		t.Run("should accept many children without RangeError", func(t *testing.T) {
			assert.Len(t,
				Filter(func(elem *dom.Node) bool { return elem.Type == dom.ElementTypeTag }, []*dom.Node{manyNodesWide}, true, math.MaxInt),
				200_001,
			)
		})

		t.Run("should turn a single node into an array", func(t *testing.T) {
			assert.Len(t,
				Filter(func(elem *dom.Node) bool { return elem.Type == dom.ElementTypeTag }, []*dom.Node{manyNodesWide.Children[0]}, true, math.MaxInt),
				200_001,
			)
		})
	})

	t.Run("findOneChild", func(t *testing.T) {
		t.Run("should find elements", func(t *testing.T) {
			assert.Equal(t,
				manyNodesWide.Children[0],
				FindOneChild(
					func(elem *dom.Node) bool { return dom.IsTag(elem) && elem.Name == "body" },
					manyNodesWide.Children,
				),
			)
		})

		t.Run("should only query direct children", func(t *testing.T) {
			assert.Nil(t,
				FindOneChild(
					func(elem *dom.Node) bool { return dom.IsTag(elem) && elem.Name == "div" },
					manyNodesWide.Children,
				),
			)
		})
	})

	t.Run("findOne", func(t *testing.T) {
		t.Run("should find elements", func(t *testing.T) {
			assert.Equal(t,
				manyNodesWide.Children[0],
				FindOne(
					func(elem *dom.Node) bool { return elem.Name == "body" },
					manyNodesWide.Children,
					true,
				),
			)
		})

		t.Run("should find elements in children", func(t *testing.T) {
			assert.Equal(t,
				manyNodesWide.Children[0].Children[0],
				FindOne(
					func(elem *dom.Node) bool { return elem.Name == "div" },
					manyNodesWide.Children,
					true,
				),
			)
		})

		t.Run("should find elements in children in any branch", func(t *testing.T) {
			assert.NotNil(t,
				FindOne(
					func(elem *dom.Node) bool { return elem.Name == "p" },
					someDeepNodes.Children,
					true,
				),
			)
		})

		t.Run("should not find elements in children if recurse is false", func(t *testing.T) {
			assert.Nil(t,
				FindOne(func(elem *dom.Node) bool { return elem.Name == "div" }, []*dom.Node{manyNodesWide}, false),
			)
		})

		t.Run("should return `null` if nothing is found", func(t *testing.T) {
			assert.Nil(t, FindOne(func(elem *dom.Node) bool { return false }, []*dom.Node{manyNodesWide}, true))
		})
	})

	t.Run("existsOne", func(t *testing.T) {
		t.Run("should find elements", func(t *testing.T) {
			assert.True(t,
				ExistsOne(func(elem *dom.Node) bool { return elem.Name == "body" }, []*dom.Node{manyNodesWide}),
			)
		})

		t.Run("should find elements in children", func(t *testing.T) {
			assert.True(t,
				ExistsOne(
					func(elem *dom.Node) bool { return elem.Name == "div" },
					manyNodesWide.Children,
				),
			)
		})

		t.Run("should return `false` if nothing is found", func(t *testing.T) {
			assert.False(t, ExistsOne(func(elem *dom.Node) bool { return false }, []*dom.Node{manyNodesWide}))
		})
	})
}
