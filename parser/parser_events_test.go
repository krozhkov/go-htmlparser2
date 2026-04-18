package parser

import (
	"log"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
)

type TestEvent struct {
	event      string
	data       []any
	endIndex   int
	startIndex int
}

type TestEventsCollectorHandler struct {
	events   []TestEvent
	parser   *Parser
	callback func(e error, events []TestEvent)
}

func NewTestEventsCollectorHandler(callback func(e error, events []TestEvent)) *TestEventsCollectorHandler {
	return &TestEventsCollectorHandler{
		events:   make([]TestEvent, 0, 10),
		parser:   nil,
		callback: callback,
	}
}
func (h *TestEventsCollectorHandler) handle(event string, data []any) {
	switch event {
	case "error":
		{
			h.callback(data[0].(error), nil)

			break
		}
	case "end":
		{
			h.callback(nil, h.events)

			break
		}
	case "reset":
		{
			h.events = h.events[:0]

			break
		}
	case "parserinit":
		{
			h.parser = data[0].(*Parser)

			// Don't collect event
			break
		}
	default:
		{
			var last *TestEvent
			if len(h.events) > 0 {
				last = &h.events[len(h.events)-1]
			}

			// Combine text nodes
			if event == "text" && last != nil && last.event == "text" {
				text := last.data[0].(string)
				text += data[0].(string)
				last.data[0] = text
				last.endIndex = h.parser.EndIndex

				break
			}

			// Remove `undefined`s from attribute responses, as they cannot be represented in JSON.
			if event == "attribute" && data[2] == QuoteTypeNoValue {
				data = data[:len(data)-1]
			}

			if h.parser.StartIndex > h.parser.EndIndex {
				log.Fatalf("Invalid start/end index %d > %d", h.parser.StartIndex, h.parser.EndIndex)
			}

			h.events = append(h.events, TestEvent{
				event:      event,
				startIndex: h.parser.StartIndex,
				endIndex:   h.parser.EndIndex,
				data:       data,
			})
		}
	}
}
func (h *TestEventsCollectorHandler) OnParserInit(parser *Parser) {
	h.handle("parserinit", []any{parser})
}
func (h *TestEventsCollectorHandler) OnReset() {
	h.handle("reset", []any{})
}
func (h *TestEventsCollectorHandler) OnEnd() {
	h.handle("end", []any{})
}
func (h *TestEventsCollectorHandler) OnError(e error) {
	h.handle("error", []any{e})
}
func (h *TestEventsCollectorHandler) OnCloseTag(name string, isImplied bool) {
	h.handle("closetag", []any{name, isImplied})
}
func (h *TestEventsCollectorHandler) OnOpenTagName(name string) {
	h.handle("opentagname", []any{name})
}
func (h *TestEventsCollectorHandler) OnAttribute(name string, value string, quote QuoteType) {
	h.handle("attribute", []any{name, value, quote})
}
func (h *TestEventsCollectorHandler) OnOpenTag(name string, attribs []*Attribute, isImplied bool) {
	h.handle("opentag", []any{name, attribs, isImplied})
}
func (h *TestEventsCollectorHandler) OnText(data string) {
	h.handle("text", []any{data})
}
func (h *TestEventsCollectorHandler) OnComment(data string) {
	h.handle("comment", []any{data})
}
func (h *TestEventsCollectorHandler) OnCDataStart() {
	h.handle("cdatastart", []any{})
}
func (h *TestEventsCollectorHandler) OnCDataEnd() {
	h.handle("cdataend", []any{})
}
func (h *TestEventsCollectorHandler) OnCommentEnd() {
	h.handle("commentend", []any{})
}
func (h *TestEventsCollectorHandler) OnProcessingInstruction(name string, data string) {
	h.handle("processinginstruction", []any{name, data})
}

/**
 * Write to the parser twice, once a bytes, once as a single blob. Then check
 * that we received the expected events.
 *
 * @internal
 * @param input Data to write.
 * @param options Parser options.
 * @returns Promise that resolves if the test passes.
 */
func runTest(t *testing.T, input string, options *ParserOptions) {
	var firstResult []TestEvent

	handler := NewTestEventsCollectorHandler(func(error error, actual []TestEvent) {
		if error != nil {
			log.Fatal(error.Error())
		}

		if firstResult != nil {
			assert.Equal(t, firstResult, actual)
		} else {
			firstResult = actual
			snaps.MatchSnapshot(t, actual)
		}
	})

	parser := NewParser(handler, options)
	// First, try to run the test via chunks
	for index := 0; index < len(input); index++ {
		parser.Write([]byte{input[index]})
	}
	parser.End(nil)
	// Then, parse everything
	parser.ParseComplete([]byte(input))
}

func TestEvents(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		runTest(t, "<h1 class=test>adsf</h1>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Template script tags", func(t *testing.T) {
		runTest(
			t,
			"<p><script type=\"text/template\"><h1>Heading1</h1></script></p>",
			&ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true},
		)
	})

	t.Run("Lowercase tags", func(t *testing.T) {
		runTest(t, "<H1 class=test>adsf</H1>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("CDATA", func(t *testing.T) {
		runTest(t, "<tag><![CDATA[ asdf ><asdf></adsf><> fo]]></tag><![CD>", &ParserOptions{XmlMode: true, DecodeEntities: true, LowerCaseAttributeNames: true})
	})

	t.Run("CDATA (inside special)", func(t *testing.T) {
		runTest(t, "<script>/*<![CDATA[*/ asdf ><asdf></adsf><> fo/*]]>*/</script>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("leading lt", func(t *testing.T) {
		runTest(t, ">a>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("end slash: void element ending with />", func(t *testing.T) {
		runTest(t, "<hr / ><p>Hold the line.", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("end slash: void element ending with >", func(t *testing.T) {
		runTest(t, "<hr   ><p>Hold the line.", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("end slash: void element ending with >, xmlMode=true", func(t *testing.T) {
		runTest(t, "<hr   ><p>Hold the line.", &ParserOptions{XmlMode: true, DecodeEntities: true, LowerCaseAttributeNames: true})
	})

	t.Run("end slash: non-void element ending with />", func(t *testing.T) {
		runTest(t, "<xx / ><p>Hold the line.", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("end slash: non-void element ending with />, xmlMode=true", func(t *testing.T) {
		runTest(t, "<xx / ><p>Hold the line.", &ParserOptions{XmlMode: true, DecodeEntities: true, LowerCaseAttributeNames: true})
	})

	t.Run("end slash: non-void element ending with />, recognizeSelfClosing=true", func(t *testing.T) {
		runTest(t, "<xx / ><p>Hold the line.", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true, RecognizeSelfClosing: true})
	})

	t.Run("end slash: as part of attrib value of void element", func(t *testing.T) {
		runTest(t, "<img src=gif.com/123/><p>Hold the line.", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("end slash: as part of attrib value of non-void element", func(t *testing.T) {
		runTest(t, "<a href=http://test.com/>Foo</a><p>Hold the line.", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Implicit close tags", func(t *testing.T) {
		runTest(t,
			"<ol><li class=test><div><table style=width:100%><tr><th>TH<td colspan=2><h3>Heading</h3><tr><td><div>Div</div><td><div>Div2</div></table></div><li><div><h3>Heading 2</h3></div></li></ol><p>Para<h4>Heading 4</h4><p><ul><li>Hi<li>bye</ul>",
			&ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true},
		)
	})

	t.Run("attributes (no white space, no value, no quotes)", func(t *testing.T) {
		runTest(t,
			"<button class=\"test0\"title=\"test1\" disabled value=test2>adsf</button>",
			&ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true},
		)
	})

	t.Run("crazy attribute", func(t *testing.T) {
		runTest(t, "<p < = '' FAIL>stuff</p><a", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Scripts creating other scripts", func(t *testing.T) {
		runTest(t, "<p><script>var str = '<script></'+'script>';</script></p>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Long comment ending", func(t *testing.T) {
		runTest(t, "<meta id='before'><!-- text ---><meta id='after'>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Long CDATA ending", func(t *testing.T) {
		runTest(t, "<before /><tag><![CDATA[ text ]]]></tag><after />", &ParserOptions{XmlMode: true, DecodeEntities: true, LowerCaseAttributeNames: true})
	})

	t.Run("Implicit open p and br tags", func(t *testing.T) {
		runTest(t, "<div>Hallo</p>World</br></ignore></div></p></br>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("lt followed by whitespace", func(t *testing.T) {
		runTest(t, "a < b", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("double attribute", func(t *testing.T) {
		runTest(t, "<h1 class=test class=boo></h1>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("numeric entities", func(t *testing.T) {
		runTest(t, "&#x61;&#x62&#99;&#100&#x66g&#x;&#x68", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("legacy entities", func(t *testing.T) {
		runTest(t, "&AMPel&iacutee&ampeer;s&lter&sum", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("named entities", func(t *testing.T) {
		runTest(t, "&amp;el&lt;er&CounterClockwiseContourIntegral;foo&bar", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("xml entities", func(t *testing.T) {
		runTest(t, "&amp;&gt;&amp&lt;&uuml;&#x61;&#x62&#99;&#100&#101", &ParserOptions{XmlMode: true, DecodeEntities: true, LowerCaseAttributeNames: true})
	})

	t.Run("entity in attribute", func(t *testing.T) {
		runTest(t,
			"<a href='http://example.com/p&#x61;#x61ge?param=value&param2&param3=&lt;val&; & &'>",
			&ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true},
		)
	})

	t.Run("double brackets", func(t *testing.T) {
		runTest(t, "<<princess-purpose>>testing</princess-purpose>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("legacy entities fail", func(t *testing.T) {
		runTest(t, "M&M", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Special special tags", func(t *testing.T) {
		runTest(t,
			"<tItLe><b>foo</b><title></TiTlE><sitle><b></b></sitle><ttyle><b></b></ttyle><sCriPT></scripter</soo</sCript><STyLE></styler</STylE><sCiPt><stylee><scriptee><soo>",
			&ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true},
		)
	})

	t.Run("Empty tag name", func(t *testing.T) {
		runTest(t, "< ></ >", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Not quite closed", func(t *testing.T) {
		runTest(t, "<foo /bar></foo bar>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Entities in attributes", func(t *testing.T) {
		runTest(t, "<foo bar=&amp; baz=\"&amp;\" boo='&amp;' noo=>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("CDATA in HTML", func(t *testing.T) {
		runTest(t, "<![CDATA[ foo ]]>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Comment edge-cases", func(t *testing.T) {
		runTest(t, "<!-foo><!-- --- --><!--foo", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("CDATA edge-cases", func(t *testing.T) {
		runTest(t, "<![CDATA><![CDATA[[]]sdaf]]><![CDATA[foo", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true, RecognizeCDATA: true})
	})

	t.Run("Comment false ending", func(t *testing.T) {
		runTest(t, "<!-- a-b-> -->", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Scripts ending with <", func(t *testing.T) {
		runTest(t, "<script><</script>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("CDATA more edge-cases", func(t *testing.T) {
		runTest(t, "<![CDATA[foo]bar]>baz]]>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true, RecognizeCDATA: true})
	})

	t.Run("tag names are not ASCII alpha", func(t *testing.T) {
		runTest(t, "<12>text</12>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("open-implies-close case of (non-br) void close tag in non-XML mode", func(t *testing.T) {
		runTest(t, "<select><input></select>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("entity in attribute (#276)", func(t *testing.T) {
		runTest(t,
			"<img src=\"?&image_uri=1&&image;=2&image=3\"/>?&image_uri=1&&image;=2&image=3",
			&ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true},
		)
	})

	t.Run("entity in title (#592)", func(t *testing.T) {
		runTest(t, "<title>the &quot;title&quot", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("entity in title - decodeEntities=false (#592)", func(t *testing.T) {
		runTest(t, "<title>the &quot;title&quot;", &ParserOptions{XmlMode: false, DecodeEntities: false, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("</title> in <script> (#745)", func(t *testing.T) {
		runTest(t, "<script>'</title>'</script>", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("XML tags", func(t *testing.T) {
		runTest(t, "<:foo><_bar>", &ParserOptions{XmlMode: true, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Trailing legacy entity", func(t *testing.T) {
		runTest(t, "&timesbar;&timesbar", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Trailing numeric entity", func(t *testing.T) {
		runTest(t, "&#53&#53", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Multi-byte entity", func(t *testing.T) {
		runTest(t, "&NotGreaterFullEqual;", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Start & end indices from domhandler", func(t *testing.T) {
		runTest(t,
			"<!DOCTYPE html> <html> <title>The Title</title> <body class='foo'>Hello world <p></p></body> <!-- the comment --> </html> ",
			&ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true},
		)
	})

	t.Run("Self-closing indices (#941)", func(t *testing.T) {
		runTest(t, "<xml><a/><b/></xml>", &ParserOptions{XmlMode: true, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Entity after <", func(t *testing.T) {
		runTest(t, "<&amp;", &ParserOptions{XmlMode: false, DecodeEntities: true, LowerCaseTags: true, LowerCaseAttributeNames: true})
	})

	t.Run("Attribute in XML (see #1350)", func(t *testing.T) {
		runTest(t,
			"<Page\n    title=\"Hello world\"\n    actionBarVisible=\"false\"/>",
			&ParserOptions{XmlMode: true, DecodeEntities: true},
		)
	})
}
