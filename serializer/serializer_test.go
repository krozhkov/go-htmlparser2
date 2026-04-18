package serializer

import (
	"testing"

	"github.com/krozhkov/go-htmlparser2/dom"
	"github.com/krozhkov/go-htmlparser2/parser"
	"github.com/stretchr/testify/assert"
)

type LoadingOptions struct {
	decodeEntities  *bool
	encodeEntities  *string
	selfClosingTags *bool
	emptyAttrs      *bool
}

func html(str string, opts *LoadingOptions) string {
	parserOpts := parser.ParserOptions{XmlMode: false, LowerCaseAttributeNames: true, DecodeEntities: true, RecognizeSelfClosing: true}
	if opts != nil && opts.decodeEntities != nil {
		parserOpts.DecodeEntities = *opts.decodeEntities
	}
	nodes := dom.ParseDOM(str, &parserOpts)
	serializerOpts := DomSerializerOptions{}
	if opts != nil && opts.decodeEntities != nil {
		serializerOpts.DecodeEntities = opts.decodeEntities
	}
	if serializerOpts.DecodeEntities != nil {
		if *serializerOpts.DecodeEntities {
			serializerOpts.EncodeEntities = ptr("true")
		} else {
			serializerOpts.EncodeEntities = ptr("false")
		}
	}
	if opts != nil && opts.encodeEntities != nil {
		serializerOpts.EncodeEntities = opts.encodeEntities
	}
	if opts != nil && opts.selfClosingTags != nil {
		serializerOpts.SelfClosingTags = opts.selfClosingTags
	}
	if opts != nil && opts.emptyAttrs != nil {
		serializerOpts.EmptyAttrs = opts.emptyAttrs
	}
	return Render(nodes, &serializerOpts)
}

func xml(str string, opts *LoadingOptions) string {
	nodes := dom.ParseDOM(str, &parser.ParserOptions{XmlMode: true})
	serializerOpts := DomSerializerOptions{XmlMode: ptr("true")}
	if opts != nil && opts.decodeEntities != nil {
		serializerOpts.DecodeEntities = opts.decodeEntities
	}
	if opts != nil && opts.selfClosingTags != nil {
		serializerOpts.SelfClosingTags = opts.selfClosingTags
	}
	return Render(nodes, &serializerOpts)
}

func TestHtmlParser2(t *testing.T) {
	t.Run("should handle double quotes within single quoted attributes properly", func(t *testing.T) {
		str := "<hr class='an \"edge\" case' />"
		assert.Equal(t, `<hr class="an &quot;edge&quot; case">`, html(str, nil))
	})

	t.Run("should escape entities to utf8 if requested", func(t *testing.T) {
		str := `<a href="a < b &quot; & c">& " &lt; &gt;</a>`
		assert.Equal(t, `<a href="a < b &quot; &amp; c">&amp; " &lt; &gt;</a>`, html(str, &LoadingOptions{encodeEntities: ptr("utf8")}))
	})

	t.Run("should render <br /> tags correctly", func(t *testing.T) {
		str := "<br />"
		assert.Equal(t, str, html(str, &LoadingOptions{decodeEntities: ptr(false), selfClosingTags: ptr(true)}))
	})

	t.Run("should render childless SVG nodes with an explicit closing tag", func(t *testing.T) {
		str := `<svg><circle x="12" y="12"></circle><path d="123M"></path><polygon points="60,20 100,40 100,80 60,100 20,80 20,40"></polygon></svg>`
		assert.Equal(t, str, html(str, &LoadingOptions{decodeEntities: ptr(false), selfClosingTags: ptr(false)}))
	})
}

func TestXml(t *testing.T) {
	t.Run("should render CDATA correctly", func(t *testing.T) {
		str := "<a> <b> <![CDATA[ asdf&asdf ]]> <c/> <![CDATA[ asdf&asdf ]]> </b> </a>"
		assert.Equal(t, str, xml(str, nil))
	})

	t.Run("should append =\"\" to attributes with no value", func(t *testing.T) {
		str := "<div dropdown-toggle>"
		assert.Equal(t, `<div dropdown-toggle=""/>`, xml(str, nil))
	})

	t.Run("should append =\"\" to boolean attributes with no value", func(t *testing.T) {
		str := "<input disabled>"
		assert.Equal(t, `<input disabled=""/>`, xml(str, nil))
	})

	t.Run("should preserve XML prefixes on attributes", func(t *testing.T) {
		str := `<div xmlns:ex="http://example.com/ns"><p ex:ample="attribute">text</p></div>`
		assert.Equal(t, str, xml(str, nil))
	})

	t.Run("should preserve mixed-case XML elements and attributes", func(t *testing.T) {
		str := `<svg viewBox="0 0 8 8"><radialGradient/></svg>`
		assert.Equal(t, str, xml(str, nil))
	})

	t.Run("should encode entities in otherwise special tags", func(t *testing.T) {
		str := `<script>"<br/>"</script>`
		assert.Equal(t, "<script>&quot;<br/>&quot;</script>", xml(str, nil))
	})

	t.Run("should not encode entities if disabled", func(t *testing.T) {
		str := `<script>"<br/>"</script>`
		assert.Equal(t, str, xml(str, &LoadingOptions{decodeEntities: ptr(false)}))
	})

	t.Run("should render childless nodes with an explicit closing tag", func(t *testing.T) {
		str := "<foo /><bar></bar>"
		assert.Equal(t, "<foo></foo><bar></bar>", xml(str, &LoadingOptions{selfClosingTags: ptr(false)}))
	})
}

func TestHtml(t *testing.T) {
	testBody(t, func(input string, opts *LoadingOptions) string {
		return html(input, opts)
	})
}

func TestHtmlDontDecodeEntities(t *testing.T) {
	testBody(t, func(input string, opts *LoadingOptions) string {
		if opts == nil {
			opts = &LoadingOptions{}
		}
		opts.decodeEntities = ptr(false)
		return html(input, opts)
	})
}

func testBody(t *testing.T, htmlFunc func(input string, opts *LoadingOptions) string) {
	t.Run("should render <br /> tags without a slash", func(t *testing.T) {
		str := "<br />"
		assert.Equal(t, "<br>", htmlFunc(str, nil))
	})

	t.Run("should retain encoded HTML content within attributes", func(t *testing.T) {
		str := `<hr class="cheerio &amp; node = happy parsing" />`
		assert.Equal(t, `<hr class="cheerio &amp; node = happy parsing">`, htmlFunc(str, nil))
	})

	t.Run(`should shorten the "checked" attribute when it contains the value "checked"`, func(t *testing.T) {
		str := "<input checked/>"
		assert.Equal(t, "<input checked>", htmlFunc(str, nil))
	})

	t.Run("should render empty attributes if asked for", func(t *testing.T) {
		str := "<input checked/>"
		assert.Equal(t, `<input checked="">`, htmlFunc(str, &LoadingOptions{emptyAttrs: ptr(true)}))
	})

	t.Run(`should not shorten the "name" attribute when it contains the value "name"`, func(t *testing.T) {
		str := `<input name="name"/>`
		assert.Equal(t, `<input name="name">`, htmlFunc(str, nil))
	})

	t.Run(`should not append ="" to attributes with no value`, func(t *testing.T) {
		str := "<div dropdown-toggle>"
		assert.Equal(t, "<div dropdown-toggle></div>", htmlFunc(str, nil))
	})

	t.Run("should render comments correctly", func(t *testing.T) {
		str := "<!-- comment -->"
		assert.Equal(t, "<!-- comment -->", htmlFunc(str, nil))
	})

	t.Run("should render whitespace by default", func(t *testing.T) {
		str := `<a href="./haha.html">hi</a> <a href="./blah.html">blah</a>`
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should preserve multiple hyphens in data attributes", func(t *testing.T) {
		str := `<div data-foo-bar-baz="value"></div>`
		assert.Equal(t, `<div data-foo-bar-baz="value"></div>`, htmlFunc(str, nil))
	})

	t.Run("should not encode characters in script tag", func(t *testing.T) {
		str := `<script>alert("hello world")</script>`
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should not encode tags in script tag", func(t *testing.T) {
		str := `<script>"<br>"</script>`
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should not encode json data", func(t *testing.T) {
		str := `<script>var json = {"simple_value": "value", "value_with_tokens": "&quot;here & \'there\'&quot;"};</script>`
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should render childless SVG nodes with a closing slash in HTML mode", func(t *testing.T) {
		str := `<svg><circle x="12" y="12"/><path d="123M"/><polygon points="60,20 100,40 100,80 60,100 20,80 20,40"/></svg>`
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should render childless MathML nodes with a closing slash in HTML mode", func(t *testing.T) {
		str := "<math><infinity/></math>"
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should allow SVG elements to have children", func(t *testing.T) {
		str := `<svg><circle cx="12" r="12"><title>dot</title></circle></svg>`
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should not include extra whitespace in SVG self-closed elements", func(t *testing.T) {
		str := `<svg><image href="x.png"/>     </svg>`
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should fix-up bad nesting in SVG in HTML mode", func(t *testing.T) {
		str := `<svg><g><image href="x.png"></svg>`
		assert.Equal(t, `<svg><g><image href="x.png"/></g></svg>`, htmlFunc(str, nil))
	})

	t.Run("should preserve XML prefixed attributes on inline SVG nodes in HTML mode", func(t *testing.T) {
		str := `<svg><text id="t" xml:lang="fr">Bonjour</text><use xlink:href="#t"/></svg>`
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should handle mixed-case SVG content in HTML mode", func(t *testing.T) {
		str := `<svg viewBox="0 0 8 8"><radialGradient/></svg>`
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should render HTML content in SVG foreignObject in HTML mode", func(t *testing.T) {
		str := `<svg><foreignObject requiredFeatures=""><img src="test.png" viewbox>text<svg viewBox="0 0 8 8"><circle r="3"/></svg></foreignObject></svg>`
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should render iframe nodes with a closing tag in HTML mode", func(t *testing.T) {
		str := `<iframe src="test"></iframe>`
		assert.Equal(t, str, htmlFunc(str, nil))
	})

	t.Run("should encode double quotes in attribute", func(t *testing.T) {
		str := `<img src="/" alt='title" onerror="alert(1)" label="x'>`
		assert.Equal(t, `<img src="/" alt="title&quot; onerror=&quot;alert(1)&quot; label=&quot;x">`, htmlFunc(str, nil))
	})
}
