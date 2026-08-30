package parser

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type HandlerMockCallbacks struct {
	OnText     func(data string)
	OnOpenTag  func(name string, attribs []*Attribute, isImplied bool)
	OnCloseTag func(name string, isImplied bool)
}

type HandlerMock struct {
	mock.Mock
	cbs HandlerMockCallbacks
}

func (m *HandlerMock) OnParserInit(parser *Parser) {
	m.Called(parser)
}
func (m *HandlerMock) OnReset() {
	m.Called()
}
func (m *HandlerMock) OnEnd() {
	m.Called()
}
func (m *HandlerMock) OnError(e error) {
	m.Called(e.Error())
}
func (m *HandlerMock) OnCloseTag(name string, isImplied bool) {
	if m.cbs.OnCloseTag != nil {
		m.cbs.OnCloseTag(name, isImplied)
	}
	m.Called(name, isImplied)
}
func (m *HandlerMock) OnOpenTagName(name string) {
	m.Called(name)
}
func (m *HandlerMock) OnAttribute(name string, value string, quote QuoteType) {
	m.Called(name, value, quote)
}
func (m *HandlerMock) OnOpenTag(name string, attribs []*Attribute, isImplied bool) {
	if m.cbs.OnOpenTag != nil {
		m.cbs.OnOpenTag(name, attribs, isImplied)
	}
	m.Called(name, fmt.Sprintf("%#v", attribs), isImplied)
}
func (m *HandlerMock) OnText(data string) {
	if m.cbs.OnText != nil {
		m.cbs.OnText(data)
	}
	m.Called(data)
}
func (m *HandlerMock) OnComment(data string) {
	m.Called(data)
}
func (m *HandlerMock) OnCDataStart() {
	m.Called()
}
func (m *HandlerMock) OnCDataEnd() {
	m.Called()
}
func (m *HandlerMock) OnCommentEnd() {
	m.Called()
}
func (m *HandlerMock) OnProcessingInstruction(name string, data string) {
	m.Called(name, data)
}
func (m *HandlerMock) ResetMock() {
	m.Calls = nil
}

func NewHandlerMock(cbs HandlerMockCallbacks) *HandlerMock {
	m := &HandlerMock{cbs: cbs}
	m.On("OnParserInit", mock.Anything).Return()
	m.On("OnReset").Return()
	m.On("OnEnd").Return()
	m.On("OnError", mock.Anything).Return()
	m.On("OnCloseTag", mock.Anything, mock.Anything).Return()
	m.On("OnOpenTagName", mock.Anything).Return()
	m.On("OnAttribute", mock.Anything, mock.Anything, mock.Anything).Return()
	m.On("OnOpenTag", mock.Anything, mock.Anything, mock.Anything).Return()
	m.On("OnText", mock.Anything).Return()
	m.On("OnComment", mock.Anything).Return()
	m.On("OnCDataStart").Return()
	m.On("OnCDataEnd").Return()
	m.On("OnCommentEnd").Return()
	m.On("OnProcessingInstruction", mock.Anything, mock.Anything).Return()
	return m
}

func TestAPI(t *testing.T) {
	writeToPipe := func(t *testing.T, w *io.PipeWriter, chunks ...string) {
		defer w.Close()
		for _, chunk := range chunks {
			_, err := w.Write([]byte(chunk))
			if err != nil {
				t.Error(err)
			}
		}
	}

	t.Run("should work without callbacks", func(t *testing.T) {
		cbs := NewHandlerMock(HandlerMockCallbacks{})

		p := NewParser(strings.NewReader("<a foo><bar></a><!-- --><![CDATA[]]]><?foo?><!bar><boo/>boohay"), cbs, &ParserOptions{
			XmlMode:                 true,
			LowerCaseAttributeNames: true,
		})

		err := p.Parse()
		require.Nil(t, err)

		err = p.Parse()
		require.Nil(t, err)

		r, w := io.Pipe()

		go writeToPipe(t, w, "<a foo", ">", "foo", "bar")

		p.Reset(r)

		err = p.Parse()
		require.Nil(t, err)

		cbs.AssertNotCalled(t, "OnText")
		cbs.AssertCalled(t, "OnText", "foo")
		cbs.AssertNumberOfCalls(t, "OnText", 4)
		cbs.AssertCalled(t, "OnText", "bar")
	})

	t.Run("should back out of numeric entities (#125)", func(t *testing.T) {
		var text string
		cbs := NewHandlerMock(HandlerMockCallbacks{OnText: func(data string) { text += data }})
		p := NewParser(strings.NewReader("id=770&#anchor"), cbs, &ParserOptions{
			XmlMode:                 false,
			LowerCaseAttributeNames: true,
		})

		err := p.Parse()
		require.Nil(t, err)

		cbs.AssertNumberOfCalls(t, "OnEnd", 1)
		assert.Equal(t, "id=770&#anchor", text)

		p.Reset(strings.NewReader("0&#xn"))
		text = ""

		err = p.Parse()
		require.Nil(t, err)

		cbs.AssertNumberOfCalls(t, "OnEnd", 2)
		assert.Equal(t, "0&#xn", text)
	})

	t.Run("should not have the start index be greater than the end index", func(t *testing.T) {
		r, w := io.Pipe()

		go writeToPipe(t, w, "<p>", "Foo", "<hr>")

		var p *Parser
		cbs := NewHandlerMock(HandlerMockCallbacks{
			OnOpenTag: func(name string, attribs []*Attribute, isImplied bool) {
				assert.LessOrEqual(t, p.StartIndex, p.EndIndex)
			},
			OnCloseTag: func(name string, isImplied bool) {
				assert.LessOrEqual(t, p.StartIndex, p.EndIndex)
			},
		})

		p = NewParser(r, cbs, &ParserOptions{
			XmlMode:                 false,
			LowerCaseAttributeNames: true,
		})

		err := p.Parse()
		require.Nil(t, err)

		cbs.AssertCalled(t, "OnOpenTag", "p", "[]*parser.Attribute{}", false)
		cbs.AssertCalled(t, "OnOpenTag", "hr", "[]*parser.Attribute{}", false)
		cbs.AssertNumberOfCalls(t, "OnCloseTag", 2)
		cbs.AssertCalled(t, "OnCloseTag", "p", true)
		cbs.AssertCalled(t, "OnCloseTag", "hr", true)
	})

	t.Run("should update the position when a single tag is spread across multiple chunks", func(t *testing.T) {
		r, w := io.Pipe()

		go writeToPipe(t, w, "<div ", "foo=bar>")

		var called bool
		var p *Parser
		cbs := NewHandlerMock(HandlerMockCallbacks{
			OnOpenTag: func(name string, attribs []*Attribute, isImplied bool) {
				called = true
				assert.Equal(t, 0, p.StartIndex)
				assert.Equal(t, 12, p.EndIndex)
			},
		})
		p = NewParser(r, cbs, &ParserOptions{
			XmlMode:                 false,
			LowerCaseAttributeNames: true,
		})

		err := p.Parse()
		require.Nil(t, err)

		assert.Equal(t, true, called)
	})

	t.Run("should have the correct position for implied opening tags", func(t *testing.T) {
		var called bool
		var p *Parser
		cbs := NewHandlerMock(HandlerMockCallbacks{
			OnOpenTag: func(name string, attribs []*Attribute, isImplied bool) {
				called = true
				assert.Equal(t, 0, p.StartIndex)
				assert.Equal(t, 3, p.EndIndex)
			},
		})
		p = NewParser(strings.NewReader("</p>"), cbs, &ParserOptions{
			XmlMode:                 false,
			LowerCaseAttributeNames: true,
		})

		err := p.Parse()
		require.Nil(t, err)

		assert.Equal(t, true, called)
	})

	t.Run("should parse <__proto__> (#387)", func(t *testing.T) {
		cbs := NewHandlerMock(HandlerMockCallbacks{})

		p := NewParser(strings.NewReader("<__proto__>"), cbs, &ParserOptions{
			XmlMode:                 false,
			LowerCaseAttributeNames: true,
		})

		// Should not throw
		e := p.Parse()
		require.Nil(t, e)

		cbs.AssertNotCalled(t, "OnError")
	})
}
