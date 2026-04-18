package domutils

import (
	"strings"

	"github.com/krozhkov/go-htmlparser2/dom"
)

/**
 * Get a node's inner text. Same as `textContent`, but inserts newlines for `<br>` tags. Ignores comments.
 *
 * @category Stringify
 * @deprecated Use `textContent` instead.
 * @param node Node to get the inner text of.
 * @returns `node`'s inner text.
 */
func GetText(node *dom.Node) string {
	var sb = new(strings.Builder)
	getText(sb, node)
	return sb.String()
}

func getText(sb *strings.Builder, node *dom.Node) {
	if dom.IsTag(node) {
		if node.Name == "br" {
			sb.WriteString("\n")
		} else {
			for _, node := range node.Children {
				getText(sb, node)
			}
		}
	}
	if dom.IsCDATA(node) {
		for _, node := range node.Children {
			getText(sb, node)
		}
	}
	if dom.IsText(node) {
		sb.WriteString(node.Data)
	}
}

/**
 * Get a node's text content. Ignores comments.
 *
 * @category Stringify
 * @param node Node to get the text content of.
 * @returns `node`'s text content.
 * @see {@link https://developer.mozilla.org/en-US/docs/Web/API/Node/textContent}
 */
func TextContent(node *dom.Node) string {
	var sb = new(strings.Builder)
	textContent(sb, node)
	return sb.String()
}

func textContent(sb *strings.Builder, node *dom.Node) {
	if dom.HasChildren(node) && !dom.IsComment(node) {
		for _, node := range node.Children {
			textContent(sb, node)
		}
	}
	if dom.IsText(node) {
		sb.WriteString(node.Data)
	}
}

/**
 * Get a node's inner text, ignoring `<script>` and `<style>` tags. Ignores comments.
 *
 * @category Stringify
 * @param node Node to get the inner text of.
 * @returns `node`'s inner text.
 * @see {@link https://developer.mozilla.org/en-US/docs/Web/API/Node/innerText}
 */
func InnerText(node *dom.Node) string {
	var sb = new(strings.Builder)
	innerText(sb, node)
	return sb.String()
}

func innerText(sb *strings.Builder, node *dom.Node) {
	if dom.HasChildren(node) && (node.Type == dom.ElementTypeTag || dom.IsCDATA(node)) {
		for _, node := range node.Children {
			innerText(sb, node)
		}
	}
	if dom.IsText(node) {
		sb.WriteString(node.Data)
	}
}
