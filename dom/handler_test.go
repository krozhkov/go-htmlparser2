package dom

import (
	"testing"
	"unicode/utf8"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/krozhkov/go-htmlparser2/parser"
	"github.com/stretchr/testify/assert"
)

type domHandlerTest struct {
	name    string
	options *DomHandlerOptions
	input   string
	output  []*Node
}

var testValues = []domHandlerTest{
	{
		name:  "Basic test",
		input: "<!DOCTYPE html><html><title>The Title</title><body>Hello world</body></html>",
		output: []*Node{
			NewProcessingInstruction("!doctype", "!DOCTYPE html"),
			NewElement("html", nil, []*Node{
				NewElement("title", nil, []*Node{
					NewText("The Title"),
				}, ""),
				NewElement("body", nil, []*Node{
					NewText("Hello world"),
				}, ""),
			}, ""),
		},
	},
	{
		name:  "Single Tag 1",
		input: "<br>text</br>",
		output: []*Node{
			NewElement("br", nil, nil, ""),
			NewText("text"),
			NewElement("br", nil, nil, ""),
		},
	},
	{
		name:  "Single Tag 2",
		input: "<br>text<br>",
		output: []*Node{
			NewElement("br", nil, nil, ""),
			NewText("text"),
			NewElement("br", nil, nil, ""),
		},
	},
	{
		name:  "Unescaped chars in script",
		input: "<head><script language=\"Javascript\">var foo = \"<bar>\"; alert(2 > foo); var baz = 10 << 2; var zip = 10 >> 1; var yap = \"<<>>>><<\";</script></head>",
		output: []*Node{
			NewElement("head", nil, []*Node{
				NewElement("script", orderedmap.NewOrderedMapWithElements(&Attribute{Key: "language", Value: "Javascript"}), []*Node{
					NewText("var foo = \"<bar>\"; alert(2 > foo); var baz = 10 << 2; var zip = 10 >> 1; var yap = \"<<>>>><<\";"),
				}, ""),
			}, ""),
		},
	},
	{
		name:  "Special char in comment",
		input: "<head><!-- commented out tags <title>Test</title>--></head>",
		output: []*Node{
			NewElement("head", nil, []*Node{
				NewComment(" commented out tags <title>Test</title>"),
			}, ""),
		},
	},
	{
		name:  "Script source in comment",
		input: "<script><!--var foo = 1;--></script>",
		output: []*Node{
			NewElement("script", nil, []*Node{
				NewText("<!--var foo = 1;-->"),
			}, ""),
		},
	},
	{
		name:  "Unescaped chars in style",
		input: "<style type=\"text/css\">\n body > p\n\t{ font-weight: bold; }</style>",
		output: []*Node{
			NewElement("style", orderedmap.NewOrderedMapWithElements(&Attribute{Key: "type", Value: "text/css"}), []*Node{
				NewText("\n body > p\n\t{ font-weight: bold; }"),
			}, ""),
		},
	},
	{
		name:  "Extra spaces in tag",
		input: "<font\t\n size='14' \n>the text</\t\nfont\t \n>",
		output: []*Node{
			NewElement("font", orderedmap.NewOrderedMapWithElements(&Attribute{Key: "size", Value: "14"}), []*Node{
				NewText("the text"),
			}, ""),
		},
	},
	{
		name:  "Unquoted attributes",
		input: "<font size= 14>the text</font>",
		output: []*Node{
			NewElement("font", orderedmap.NewOrderedMapWithElements(&Attribute{Key: "size", Value: "14"}), []*Node{
				NewText("the text"),
			}, ""),
		},
	},
	{
		name:  "Singular attribute",
		input: "<option value='foo' selected>",
		output: []*Node{
			NewElement("option", orderedmap.NewOrderedMapWithElements(&Attribute{Key: "value", Value: "foo"}, &Attribute{Key: "selected", Value: ""}), nil, ""),
		},
	},
	{
		name:  "Text outside tags",
		input: "Line one\n<br>\nline two",
		output: []*Node{
			NewText("Line one\n"),
			NewElement("br", nil, nil, ""),
			NewText("\nline two"),
		},
	},
	{
		name:  "Only text",
		input: "this is the text",
		output: []*Node{
			NewText("this is the text"),
		},
	},
	{
		name:  "Comment within text",
		input: "this is <!-- the comment --> the text",
		output: []*Node{
			NewText("this is "),
			NewComment(" the comment "),
			NewText(" the text"),
		},
	},
	{
		name:  "Comment within text within script",
		input: "<script>this is <!-- the comment --> the text</script>",
		output: []*Node{
			NewElement("script", nil, []*Node{
				NewText("this is <!-- the comment --> the text"),
			}, ""),
		},
	},
	{
		name:  "Option 'verbose' set to 'false'",
		input: "<font\t\n size='14' \n>the text</\t\nfont\t \n>",
		output: []*Node{
			NewElement("font", orderedmap.NewOrderedMapWithElements(&Attribute{Key: "size", Value: "14"}), []*Node{
				NewText("the text"),
			}, ""),
		},
	},
	{
		name:  "XML Namespace",
		input: "<ns:tag>text</ns:tag>",
		output: []*Node{
			NewElement("ns:tag", nil, []*Node{
				NewText("text"),
			}, ""),
		},
	},
	{
		name:  "Enforce empty tags",
		input: "<link>text</link>",
		output: []*Node{
			NewElement("link", nil, nil, ""),
			NewText("text"),
		},
	},
	{
		name:    "Ignore empty tags (xml mode)",
		input:   "<link>text</link>",
		options: &DomHandlerOptions{XmlMode: true},
		output: []*Node{
			NewElement("link", nil, []*Node{
				NewText("text"),
			}, ""),
		},
	},
	{
		name:  "Template script tags",
		input: "<script type=\"text/template\"><h1>Heading1</h1></script>",
		output: []*Node{
			NewElement("script", orderedmap.NewOrderedMapWithElements(&Attribute{Key: "type", Value: "text/template"}), []*Node{
				NewText("<h1>Heading1</h1>"),
			}, ""),
		},
	},
	{
		name:  "Conditional comments",
		input: "<!--[if lt IE 7]> <html class='no-js ie6 oldie' lang='en'> <![endif]--><!--[if lt IE 7]> <html class='no-js ie6 oldie' lang='en'> <![endif]-->",
		output: []*Node{
			NewComment("[if lt IE 7]> <html class='no-js ie6 oldie' lang='en'> <![endif]"),
			NewComment("[if lt IE 7]> <html class='no-js ie6 oldie' lang='en'> <![endif]"),
		},
	},
	{
		name:  "lowercase tags",
		input: "<!DOCTYPE html><HTML><TITLE>The Title</title><BODY>Hello world</body></html>",
		output: []*Node{
			NewProcessingInstruction("!doctype", "!DOCTYPE html"),
			NewElement("html", nil, []*Node{
				NewElement("title", nil, []*Node{
					NewText("The Title"),
				}, ""),
				NewElement("body", nil, []*Node{
					NewText("Hello world"),
				}, ""),
			}, ""),
		},
	},
	{
		name:  "DOM level 1",
		input: "<div>some stray text<h1>Hello, world.</h1><!-- comment node -->more stray text</div>",
		output: []*Node{
			NewElement("div", nil, []*Node{
				NewText("some stray text"),
				NewElement("h1", nil, []*Node{
					NewText("Hello, world."),
				}, ""),
				NewComment(" comment node "),
				NewText("more stray text"),
			}, ""),
		},
	},
	{
		name:    "withStartIndices adds correct startIndex properties",
		input:   "<!DOCTYPE html> <html> <title>The Title</title> <body class='foo'>Hello world <p></p></body> <!-- the comment --> </html>",
		options: &DomHandlerOptions{WithStartIndices: true},
		output: []*Node{
			{
				Type: "directive", NodeType: 1, StartIndex: 0,
				Data: "!DOCTYPE html",
				Name: "!doctype",
			},
			{
				Type: "text", NodeType: 3, StartIndex: 15,
				Data: " ",
			},
			{
				Type: "tag", NodeType: 1, StartIndex: 16,
				Children: []*Node{
					{
						Type: "text", NodeType: 3, StartIndex: 22,
						Data: " ",
					},
					{
						Type: "tag", NodeType: 1, StartIndex: 23,
						Children: []*Node{
							{
								Type: "text", NodeType: 3, StartIndex: 30,
								Data: "The Title",
							},
						},
						Name: "title",
					},
					{
						Type: "text", NodeType: 3, StartIndex: 47,
						Data: " ",
					},
					{
						Type: "tag", NodeType: 1, StartIndex: 48,
						Children: []*Node{
							{
								Type: "text", NodeType: 3, StartIndex: 66,
								Data: "Hello world ",
							},
							{
								Type: "tag", NodeType: 1, StartIndex: 78,
								Name: "p",
							},
						},
						Attribs: orderedmap.NewOrderedMapWithElements(&Attribute{Key: "class", Value: "foo"}),
						Name:    "body",
					},
					{
						Type: "text", NodeType: 3, StartIndex: 92,
						Data: " ",
					},
					{
						Type: "comment", NodeType: 8, StartIndex: 93,
						Data: " the comment ",
					},
					{
						Type: "text", NodeType: 3, StartIndex: 113,
						Data: " ",
					},
				},
				Name: "html",
			},
		},
	},
	{
		name:    "withEndIndices adds correct endIndex properties",
		input:   "<!DOCTYPE html> <html> <title>The Title</title> <body class='foo'>Hello world <p></p></body> <!-- the comment --> </html>",
		options: &DomHandlerOptions{WithEndIndices: true},
		output: []*Node{
			{
				Type: "directive", NodeType: 1, EndIndex: 14,
				Data: "!DOCTYPE html",
				Name: "!doctype",
			},
			{
				Type: "text", NodeType: 3, EndIndex: 15,
				Data: " ",
			},
			{
				Type: "tag", NodeType: 1, EndIndex: 120,
				Children: []*Node{
					{
						Type: "text", NodeType: 3, EndIndex: 22,
						Data: " ",
					},
					{
						Type: "tag", NodeType: 1, EndIndex: 46,
						Children: []*Node{
							{
								Type: "text", NodeType: 3, EndIndex: 38,
								Data: "The Title",
							},
						},
						Name: "title",
					},
					{
						Type: "text", NodeType: 3, EndIndex: 47,
						Data: " ",
					},
					{
						Type: "tag", NodeType: 1, EndIndex: 91,
						Children: []*Node{
							{
								Type: "text", NodeType: 3, EndIndex: 77,
								Data: "Hello world ",
							},
							{
								Type: "tag", NodeType: 1, EndIndex: 84,
								Name: "p",
							},
						},
						Attribs: orderedmap.NewOrderedMapWithElements(&Attribute{Key: "class", Value: "foo"}),
						Name:    "body",
					},
					{
						Type: "text", NodeType: 3, EndIndex: 92,
						Data: " ",
					},
					{
						Type: "comment", NodeType: 8, EndIndex: 112,
						Data: " the comment ",
					},
					{
						Type: "text", NodeType: 3, EndIndex: 113,
						Data: " ",
					},
				},
				Name: "html",
			},
		},
	},
	{
		name:  "Root-level text",
		input: "<em>hello</em> world",
		output: []*Node{
			NewElement("em", nil, []*Node{
				NewText("hello"),
			}, ""),
			NewText(" world"),
		},
	},
	{
		name:    "XML mode: All tags should have `type: tag`",
		input:   "<script><style><div>",
		options: &DomHandlerOptions{XmlMode: true},
		output: []*Node{
			{
				Type: "tag", NodeType: 1,
				Children: []*Node{
					{
						Type: "tag", NodeType: 1,
						Children: []*Node{
							{
								Type: "tag", NodeType: 1,
								Name: "div",
							},
						},
						Name: "style",
					},
				},
				Name: "script",
			},
		},
	},
}

func parseDh(t *testing.T, html string, dhOptions *DomHandlerOptions) []*Node {
	handler := NewDomHandler(func(err error, _ []*Node) {
		assert.Nil(t, err)
	}, dhOptions, nil)

	var xmlMode bool
	if dhOptions != nil {
		xmlMode = dhOptions.XmlMode
	}

	pOptions := &parser.ParserOptions{
		XmlMode:       xmlMode,
		LowerCaseTags: !xmlMode,
	}

	parser := parser.NewParser(handler, pOptions)

	for _, cp := range html {
		parser.Write(utf8.AppendRune([]byte{}, cp))
	}

	parser.End(nil)

	byChunks := prepareNode(handler.Root)

	parser.Reset()
	parser.End([]byte(html))

	result := prepareNode(handler.Root)

	assert.Equal(t, byChunks, result)

	children := result.Children

	return children
}

func TestDomHandler(t *testing.T) {
	for _, tt := range testValues {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDh(t, tt.input, tt.options)

			assert.NotNil(t, result)
			assert.Equal(t, tt.output, result)
		})
	}
}
