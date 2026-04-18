package entities

import (
	"fmt"
	"strings"
)

/**
 * Encodes all non-ASCII characters, as well as characters not valid in XML
 * documents using XML entities.
 *
 * If a character has no equivalent entity, a
 * numeric hexadecimal reference (eg. `&#xfc;`) will be used.
 */
func EncodeXML(input string) string {
	sb := new(strings.Builder)

	for _, r := range input {
		switch {
		case r == '"':
			sb.WriteString("&quot;")
		case r == '&':
			sb.WriteString("&amp;")
		case r == '\'':
			sb.WriteString("&apos;")
		case r == '<':
			sb.WriteString("&lt;")
		case r == '>':
			sb.WriteString("&gt;")
		case r >= 0x0080 && r <= 0x10FFFF:
			{
				sb.WriteString("&#x")
				sb.WriteString(fmt.Sprintf("%x", r))
				sb.WriteString(";")
			}
		default:
			sb.WriteRune(r)
		}
	}

	return sb.String()
}

/**
 * Encodes all characters not valid in XML documents using XML entities.
 *
 * Note that the output will be character-set dependent.
 *
 * @param data String to escape.
 */
func EscapeUTF8(input string) string {
	sb := new(strings.Builder)

	for _, r := range input {
		switch {
		case r == '"':
			sb.WriteString("&quot;")
		case r == '&':
			sb.WriteString("&amp;")
		case r == '\'':
			sb.WriteString("&apos;")
		case r == '<':
			sb.WriteString("&lt;")
		case r == '>':
			sb.WriteString("&gt;")
		default:
			sb.WriteRune(r)
		}
	}

	return sb.String()
}

/**
 * Encodes all characters that have to be escaped in HTML attributes,
 * following {@link https://html.spec.whatwg.org/multipage/parsing.html#escapingString}.
 *
 * @param data String to escape.
 */
func EscapeAttribute(input string) string {
	sb := new(strings.Builder)

	for _, r := range input {
		switch {
		case r == '"':
			sb.WriteString("&quot;")
		case r == '&':
			sb.WriteString("&amp;")
		case r == 0x00A0:
			sb.WriteString("&nbsp;")
		default:
			sb.WriteRune(r)
		}
	}

	return sb.String()
}

/**
 * Encodes all characters that have to be escaped in HTML text,
 * following {@link https://html.spec.whatwg.org/multipage/parsing.html#escapingString}.
 *
 * @param data String to escape.
 */
func EscapeText(input string) string {
	sb := new(strings.Builder)

	for _, r := range input {
		switch {
		case r == '&':
			sb.WriteString("&amp;")
		case r == '<':
			sb.WriteString("&lt;")
		case r == '>':
			sb.WriteString("&gt;")
		case r == 0x00A0:
			sb.WriteString("&nbsp;")
		default:
			sb.WriteRune(r)
		}
	}

	return sb.String()
}
