package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	t.Run("should work without callbacks", func(t *testing.T) {
		cbs := NewHandlerMock(HandlerMockCallbacks{})

		p := NewParser(cbs, &ParserOptions{
			XmlMode:                 true,
			LowerCaseAttributeNames: true,
		})

		p.End([]byte("<a foo><bar></a><!-- --><![CDATA[]]]><?foo?><!bar><boo/>boohay"))
		p.Write([]byte("foo"))

		// Check for an error
		p.End(nil)
		p.Write([]byte("foo"))
		cbs.AssertCalled(t, "OnError", ".write() after done")
		p.End(nil)
		cbs.AssertCalled(t, "OnError", ".end() after done")

		p.Write([]byte("foo"))
		p.Reset()

		// Remove method
		p.Write([]byte("<a foo"))
		p.Write([]byte(">"))

		cbs.ResetMock()
		// Pause/resume
		p.Pause()
		p.Write([]byte("foo"))
		cbs.AssertNotCalled(t, "OnText")
		p.Resume()
		cbs.AssertCalled(t, "OnText", "foo")
		p.Pause()
		cbs.AssertNumberOfCalls(t, "OnText", 1)
		p.Resume()
		cbs.AssertNumberOfCalls(t, "OnText", 1)
		p.Pause()
		p.End([]byte("bar"))
		cbs.AssertNumberOfCalls(t, "OnText", 1)
		p.Resume()
		cbs.AssertNumberOfCalls(t, "OnText", 2)
		cbs.AssertCalled(t, "OnText", "bar")
	})

	t.Run("should back out of numeric entities (#125)", func(t *testing.T) {
		var text string
		cbs := NewHandlerMock(HandlerMockCallbacks{OnText: func(data string) { text += data }})
		p := NewParser(cbs, &ParserOptions{
			XmlMode:                 false,
			LowerCaseAttributeNames: true,
		})

		p.End([]byte("id=770&#anchor"))

		cbs.AssertNumberOfCalls(t, "OnEnd", 1)
		assert.Equal(t, "id=770&#anchor", text)

		p.Reset()
		text = ""

		p.End([]byte("0&#xn"))

		cbs.AssertNumberOfCalls(t, "OnEnd", 2)
		assert.Equal(t, "0&#xn", text)
	})

	t.Run("should not have the start index be greater than the end index", func(t *testing.T) {
		var p *Parser
		cbs := NewHandlerMock(HandlerMockCallbacks{
			OnOpenTag: func(name string, attribs []*Attribute, isImplied bool) {
				assert.LessOrEqual(t, p.StartIndex, p.EndIndex)
			},
			OnCloseTag: func(name string, isImplied bool) {
				assert.LessOrEqual(t, p.StartIndex, p.EndIndex)
			},
		})

		p = NewParser(cbs, &ParserOptions{
			XmlMode:                 false,
			LowerCaseAttributeNames: true,
		})

		p.Write([]byte("<p>"))

		cbs.AssertCalled(t, "OnOpenTag", "p", "[]*parser.Attribute{}", false)
		cbs.AssertNotCalled(t, "OnCloseTag")

		p.Write([]byte("Foo"))

		p.Write([]byte("<hr>"))

		cbs.AssertCalled(t, "OnOpenTag", "hr", "[]*parser.Attribute{}", false)
		cbs.AssertNumberOfCalls(t, "OnCloseTag", 2)
		cbs.AssertCalled(t, "OnCloseTag", "p", true)
		cbs.AssertCalled(t, "OnCloseTag", "hr", true)
	})

	t.Run("should update the position when a single tag is spread across multiple chunks", func(t *testing.T) {
		var called bool
		var p *Parser
		cbs := NewHandlerMock(HandlerMockCallbacks{
			OnOpenTag: func(name string, attribs []*Attribute, isImplied bool) {
				called = true
				assert.Equal(t, 0, p.StartIndex)
				assert.Equal(t, 12, p.EndIndex)
			},
		})
		p = NewParser(cbs, &ParserOptions{
			XmlMode:                 false,
			LowerCaseAttributeNames: true,
		})

		p.Write([]byte("<div "))
		p.Write([]byte("foo=bar>"))

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
		p = NewParser(cbs, &ParserOptions{
			XmlMode:                 false,
			LowerCaseAttributeNames: true,
		})

		p.Write([]byte("</p>"))
		assert.Equal(t, true, called)
	})

	t.Run("should parse <__proto__> (#387)", func(t *testing.T) {
		cbs := NewHandlerMock(HandlerMockCallbacks{})

		p := NewParser(cbs, &ParserOptions{
			XmlMode:                 false,
			LowerCaseAttributeNames: true,
		})

		// Should not throw
		p.Write([]byte("<__proto__>"))

		cbs.AssertNotCalled(t, "OnError")
	})
}
