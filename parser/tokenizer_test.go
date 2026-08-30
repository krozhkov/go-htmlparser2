package parser

import (
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
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

func tokenize(t *testing.T, data string, options TokenizerOptions) [][]any {
	ttc := NewTestTokenizerCallbacks()
	tokenizer := NewTokenizer(
		strings.NewReader(data),
		options,
		ttc,
	)

	err := tokenizer.Parse()
	require.Nil(t, err)

	return ttc.log
}

func TestTokenizer(t *testing.T) {
	t.Run("should support self-closing special tags", func(t *testing.T) {
		t.Run("for self-closing script tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<script /><div></div>", TokenizerOptions{}))
		})
		t.Run("for self-closing style tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<style /><div></div>", TokenizerOptions{}))
		})
		t.Run("for self-closing title tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<title /><div></div>", TokenizerOptions{}))
		})
		t.Run("for self-closing textarea tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<textarea /><div></div>", TokenizerOptions{}))
		})
		t.Run("for self-closing xmp tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<xmp /><div></div>", TokenizerOptions{}))
		})
	})

	t.Run("should support standard special tags", func(t *testing.T) {
		t.Run("for normal script tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<script></script><div></div>", TokenizerOptions{}))
		})
		t.Run("for normal style tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<style></style><div></div>", TokenizerOptions{}))
		})
		t.Run("for normal sitle tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<title></title><div></div>", TokenizerOptions{}))
		})
		t.Run("for normal textarea tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<textarea></textarea><div></div>", TokenizerOptions{}))
		})
		t.Run("for normal xmp tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<xmp></xmp><div></div>", TokenizerOptions{}))
		})
	})

	t.Run("should treat html inside special tags as text", func(t *testing.T) {
		t.Run("for div inside script tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<script><div></div></script>", TokenizerOptions{}))
		})
		t.Run("for div inside style tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<style><div></div></style>", TokenizerOptions{}))
		})
		t.Run("for div inside title tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<title><div></div></title>", TokenizerOptions{}))
		})
		t.Run("for div inside textarea tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<textarea><div></div></textarea>", TokenizerOptions{}))
		})
		t.Run("for div inside xmp tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<xmp><div></div></xmp>", TokenizerOptions{}))
		})
	})

	t.Run("should correctly mark attributes", func(t *testing.T) {
		t.Run("for no value attribute", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<div aaaaaaa >", TokenizerOptions{}))
		})
		t.Run("for no quotes attribute", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<div aaa=aaa >", TokenizerOptions{}))
		})
		t.Run("for single quotes attribute", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<div aaa='a' >", TokenizerOptions{}))
		})
		t.Run("for double quotes attribute", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<div aaa=\"a\" >", TokenizerOptions{}))
		})
	})

	t.Run("should not break after special tag followed by an entity", func(t *testing.T) {
		t.Run("for normal special tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<style>a{}</style>&apos;<br/>", TokenizerOptions{DecodeEntities: true}))
		})
		t.Run("for self-closing special tag", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<style />&apos;<br/>", TokenizerOptions{DecodeEntities: true}))
		})
	})

	t.Run("should handle entities", func(t *testing.T) {
		t.Run("for XML entities", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "&amp;&gt;&amp&lt;&uuml;&#x61;&#x62&#99;&#100&#101", TokenizerOptions{DecodeEntities: true, XmlMode: true}))
		})

		t.Run("for entities in attributes (#276)", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "<img src=\"?&image_uri=1&&image;=2&image=3\"/>?&image_uri=1&&image;=2&image=3", TokenizerOptions{DecodeEntities: true}))
		})

		t.Run("for trailing legacy entity", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "&timesbar;&timesbar", TokenizerOptions{DecodeEntities: true}))
		})

		t.Run("for multi-byte entities", func(t *testing.T) {
			snaps.MatchSnapshot(t, tokenize(t, "&NotGreaterFullEqual;", TokenizerOptions{DecodeEntities: true}))
		})
	})
}
