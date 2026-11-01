package social

import (
	"net/url"
	"regexp"
	"strings"
)

// Node is one element of a parsed review (SDD 14 §12.10, open issue 10). The server parses the Markdown subset into
// this tree and React renders it node by node, so user text only ever reaches the page as escaped text and no HTML
// string is injected. T is one of p, br, strong, em, code, a, ul, ol, li, blockquote, text.
type Node struct {
	T    string `json:"t"`
	Text string `json:"text,omitempty"`
	Href string `json:"href,omitempty"`
	C    []Node `json:"c,omitempty"`
}

// maxDepth bounds blockquote and inline nesting; deeper markers stay literal text.
const maxDepth = 4

var (
	ulItem = regexp.MustCompile(`^[-*] +`)
	olItem = regexp.MustCompile(`^[0-9]{1,9}\. +`)
)

// Markdown parses the review subset: bold, italic, inline code, http(s) links, unordered and ordered lists,
// blockquotes and line breaks. Raw HTML, images, headings, tables and bare URLs stay plain text.
func Markdown(src string) []Node {
	return blocks(strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n"), 0)
}

func blocks(lines []string, depth int) []Node {
	out := []Node{}

	for i := 0; i < len(lines); {
		line := strings.TrimRight(lines[i], " \t")

		switch {
		case strings.TrimSpace(line) == "":
			i++
		case depth < maxDepth && strings.HasPrefix(line, ">"):
			var inner []string
			for ; i < len(lines) && strings.HasPrefix(lines[i], ">"); i++ {
				inner = append(inner, strings.TrimPrefix(strings.TrimPrefix(lines[i], ">"), " "))
			}

			out = append(out, Node{T: "blockquote", C: blocks(inner, depth+1)})
		case ulItem.MatchString(line), olItem.MatchString(line):
			marker, kind := ulItem, "ul"
			if !ulItem.MatchString(line) {
				marker, kind = olItem, "ol"
			}

			list := Node{T: kind}
			for ; i < len(lines) && marker.MatchString(lines[i]); i++ {
				item := strings.TrimRight(marker.ReplaceAllString(lines[i], ""), " \t")
				list.C = append(list.C, Node{T: "li", C: inline(item, depth)})
			}

			out = append(out, list)
		default:
			p := Node{T: "p"}
			for ; i < len(lines) && strings.TrimSpace(lines[i]) != "" && !startsBlock(lines[i], depth); i++ {
				if len(p.C) > 0 {
					p.C = append(p.C, Node{T: "br"})
				}

				p.C = append(p.C, inline(strings.TrimRight(lines[i], " \t"), depth)...)
			}

			out = append(out, p)
		}
	}

	return out
}

func startsBlock(line string, depth int) bool {
	return (depth < maxDepth && strings.HasPrefix(line, ">")) || ulItem.MatchString(line) || olItem.MatchString(line)
}

// inline parses emphasis, code and links; an opener without its closer is literal text.
func inline(s string, depth int) []Node {
	var (
		out  []Node
		text strings.Builder
	)

	flush := func() {
		if text.Len() > 0 {
			out = append(out, Node{T: "text", Text: text.String()})
			text.Reset()
		}
	}

	for i := 0; i < len(s); {
		rest := s[i:]

		if rest[0] == '`' {
			if end := strings.IndexByte(rest[1:], '`'); end > 0 {
				flush()
				out = append(out, Node{T: "code", Text: rest[1 : 1+end]})
				i += end + 2

				continue
			}
		}

		if depth < maxDepth && strings.HasPrefix(rest, "**") {
			if end := strings.Index(rest[2:], "**"); end > 0 {
				flush()
				out = append(out, Node{T: "strong", C: inline(rest[2:2+end], depth+1)})
				i += end + 4

				continue
			}

			// An unclosed "**" is literal; it must not reopen as two single-star emphases.
			text.WriteString("**")
			i += 2

			continue
		}

		if depth < maxDepth && (rest[0] == '*' || rest[0] == '_') && !strings.HasPrefix(rest, "**") {
			if end := strings.IndexByte(rest[1:], rest[0]); end > 0 {
				flush()
				out = append(out, Node{T: "em", C: inline(rest[1:1+end], depth+1)})
				i += end + 2

				continue
			}
		}

		// "![" is an image: not in the subset, so the bang keeps the bracket from opening a link.
		if rest[0] == '[' && (i == 0 || s[i-1] != '!') {
			if n, href, label, ok := link(rest); ok {
				flush()
				out = append(out, Node{T: "a", Href: href, C: inline(label, maxDepth)})
				i += n

				continue
			}
		}

		text.WriteByte(rest[0])
		i++
	}

	flush()

	return out
}

// link matches "[label](url)" at the start of s with an absolute http(s) URL and returns its length.
func link(s string) (n int, href, label string, ok bool) {
	mid := strings.Index(s, "](")
	if mid <= 1 {
		return 0, "", "", false
	}

	end := strings.IndexByte(s[mid+2:], ')')
	if end <= 0 {
		return 0, "", "", false
	}

	raw := s[mid+2 : mid+2+end]

	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || strings.ContainsAny(raw, " \t") {
		return 0, "", "", false
	}

	return mid + 3 + end, u.String(), s[1:mid], true
}
