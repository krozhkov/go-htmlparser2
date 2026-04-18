package parser

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
)

type TestTokenizerCallbacks struct {
	log [][]any
}

func NewTestTokenizerCallbacks() *TestTokenizerCallbacks {
	return &TestTokenizerCallbacks{log: make([][]any, 0, 10)}
}
func (ttc *TestTokenizerCallbacks) OnAttribData(start, endIndex int) {
	ttc.log = append(ttc.log, []any{"OnAttribData", start, endIndex})
}
func (ttc *TestTokenizerCallbacks) OnAttribEntity(codepoint rune) {
	ttc.log = append(ttc.log, []any{"OnAttribEntity", codepoint})
}
func (ttc *TestTokenizerCallbacks) OnAttribEnd(quote QuoteType, endIndex int) {
	ttc.log = append(ttc.log, []any{"OnAttribEnd", quote, endIndex})
}
func (ttc *TestTokenizerCallbacks) OnAttribName(start, endIndex int) {
	ttc.log = append(ttc.log, []any{"OnAttribName", start, endIndex})
}
func (ttc *TestTokenizerCallbacks) OnCData(start, endIndex, endOffset int) {
	ttc.log = append(ttc.log, []any{"OnCData", start, endIndex, endOffset})
}
func (ttc *TestTokenizerCallbacks) OnCloseTag(start, endIndex int) {
	ttc.log = append(ttc.log, []any{"OnCloseTag", start, endIndex})
}
func (ttc *TestTokenizerCallbacks) OnComment(start, endIndex, endOffset int) {
	ttc.log = append(ttc.log, []any{"OnComment", start, endIndex, endOffset})
}
func (ttc *TestTokenizerCallbacks) OnDeclaration(start, endIndex int) {
	ttc.log = append(ttc.log, []any{"OnDeclaration", start, endIndex})
}
func (ttc *TestTokenizerCallbacks) OnEnd() { ttc.log = append(ttc.log, []any{"OnEnd"}) }
func (ttc *TestTokenizerCallbacks) OnOpenTagEnd(endIndex int) {
	ttc.log = append(ttc.log, []any{"OnOpenTagEnd", endIndex})
}
func (ttc *TestTokenizerCallbacks) OnOpenTagName(start, endIndex int) {
	ttc.log = append(ttc.log, []any{"OnOpenTagName", start, endIndex})
}
func (ttc *TestTokenizerCallbacks) OnProcessingInstruction(start, endIndex int) {
	ttc.log = append(ttc.log, []any{"OnProcessingInstruction", start, endIndex})
}
func (ttc *TestTokenizerCallbacks) OnSelfClosingTag(endIndex int) {
	ttc.log = append(ttc.log, []any{"OnSelfClosingTag", endIndex})
}
func (ttc *TestTokenizerCallbacks) OnText(start, endIndex int) {
	ttc.log = append(ttc.log, []any{"OnText", start, endIndex})
}
func (ttc *TestTokenizerCallbacks) OnTextEntity(codepoint rune, endIndex int) {
	ttc.log = append(ttc.log, []any{"OnTextEntity", codepoint, endIndex})
}

func tokenize(data string, options TokenizerOptions) [][]any {
	ttc := NewTestTokenizerCallbacks()
	tokenizer := NewTokenizer(
		options,
		ttc,
	)

	tokenizer.Write([]byte(data))
	tokenizer.End()

	return ttc.log
}

func TestTokenizer(t *testing.T) {
	t.Run("should support self-closing special tags", func(t *testing.T) {
		t.Run("for self-closing script tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<script /><div></div>", TokenizerOptions{}))
		})
		t.Run("for self-closing style tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<style /><div></div>", TokenizerOptions{}))
		})
		t.Run("for self-closing title tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<title /><div></div>", TokenizerOptions{}))
		})
		t.Run("for self-closing textarea tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<textarea /><div></div>", TokenizerOptions{}))
		})
		t.Run("for self-closing xmp tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<xmp /><div></div>", TokenizerOptions{}))
		})
	})

	t.Run("should support standard special tags", func(t *testing.T) {
		t.Run("for normal script tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<script></script><div></div>", TokenizerOptions{}))
		})
		t.Run("for normal style tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<style></style><div></div>", TokenizerOptions{}))
		})
		t.Run("for normal sitle tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<title></title><div></div>", TokenizerOptions{}))
		})
		t.Run("for normal textarea tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<textarea></textarea><div></div>", TokenizerOptions{}))
		})
		t.Run("for normal xmp tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<xmp></xmp><div></div>", TokenizerOptions{}))
		})
	})

	t.Run("should treat html inside special tags as text", func(t *testing.T) {
		t.Run("for div inside script tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<script><div></div></script>", TokenizerOptions{}))
		})
		t.Run("for div inside style tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<style><div></div></style>", TokenizerOptions{}))
		})
		t.Run("for div inside title tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<title><div></div></title>", TokenizerOptions{}))
		})
		t.Run("for div inside textarea tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<textarea><div></div></textarea>", TokenizerOptions{}))
		})
		t.Run("for div inside xmp tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<xmp><div></div></xmp>", TokenizerOptions{}))
		})
	})

	t.Run("should correctly mark attributes", func(t *testing.T) {
		t.Run("for no value attribute", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<div aaaaaaa >", TokenizerOptions{}))
		})
		t.Run("for no quotes attribute", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<div aaa=aaa >", TokenizerOptions{}))
		})
		t.Run("for single quotes attribute", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<div aaa='a' >", TokenizerOptions{}))
		})
		t.Run("for double quotes attribute", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<div aaa=\"a\" >", TokenizerOptions{}))
		})
	})

	t.Run("should not break after special tag followed by an entity", func(t *testing.T) {
		t.Run("for normal special tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<style>a{}</style>&apos;<br/>", TokenizerOptions{DecodeEntities: true}))
		})
		t.Run("for self-closing special tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<style />&apos;<br/>", TokenizerOptions{DecodeEntities: true}))
		})
	})

	t.Run("should handle entities", func(t *testing.T) {
		t.Run("for XML entities", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("&amp;&gt;&amp&lt;&uuml;&#x61;&#x62&#99;&#100&#101", TokenizerOptions{DecodeEntities: true, XmlMode: true}))
		})

		t.Run("for entities in attributes (#276)", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("<img src=\"?&image_uri=1&&image;=2&image=3\"/>?&image_uri=1&&image;=2&image=3", TokenizerOptions{DecodeEntities: true}))
		})

		t.Run("for trailing legacy entity", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("&timesbar;&timesbar", TokenizerOptions{DecodeEntities: true}))
		})

		t.Run("for multi-byte entities", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize("&NotGreaterFullEqual;", TokenizerOptions{DecodeEntities: true}))
		})
	})

	t.Run("should not lose data when pausing", func(t *testing.T) {
		ttc := NewTestTokenizerCallbacks()
		tokenizer := NewTokenizer(
			TokenizerOptions{DecodeEntities: true},
			ttc,
		)

		tokenizer.Write([]byte("&am"))
		tokenizer.Write([]byte("p; it up!"))
		tokenizer.Resume()
		tokenizer.Resume()

		assert.True(t, tokenizer.running, "Tokenizer shouldn't be paused")

		tokenizer.End()

		snaps.MatchSnapshot(t, ttc.log)
	})
}
