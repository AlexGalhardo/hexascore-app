package social

import (
	"encoding/json"
	"strings"
	"testing"
)

func render(t *testing.T, src string) string {
	t.Helper()

	raw, err := json.Marshal(Markdown(src))
	if err != nil {
		t.Fatal(err)
	}

	return string(raw)
}

func TestMarkdownSubset(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{
			"script is text", "<script>alert(1)</script>",
			`[{"t":"p","c":[{"t":"text","text":"\u003cscript\u003ealert(1)\u003c/script\u003e"}]}]`,
		},
		{
			"bold italic code", "**b** *i* _j_ `x*y`",
			`[{"t":"p","c":[{"t":"strong","c":[{"t":"text","text":"b"}]},{"t":"text","text":" "},` +
				`{"t":"em","c":[{"t":"text","text":"i"}]},{"t":"text","text":" "},{"t":"em","c":[{"t":"text","text":"j"}]},` +
				`{"t":"text","text":" "},{"t":"code","text":"x*y"}]}]`,
		},
		{
			"https link", "[site](https://example.com/a?b=1)",
			`[{"t":"p","c":[{"t":"a","href":"https://example.com/a?b=1","c":[{"t":"text","text":"site"}]}]}]`,
		},
		{
			"javascript link is text", "[x](javascript:alert(1))",
			`[{"t":"p","c":[{"t":"text","text":"[x](javascript:alert(1))"}]}]`,
		},
		{"relative link is text", "[x](/admin)", `[{"t":"p","c":[{"t":"text","text":"[x](/admin)"}]}]`},
		{
			"image is text", "![a](https://example.com/i.png)",
			`[{"t":"p","c":[{"t":"text","text":"![a](https://example.com/i.png)"}]}]`,
		},
		{
			"bare url is text",
			"see https://example.com",
			`[{"t":"p","c":[{"t":"text","text":"see https://example.com"}]}]`,
		},
		{"heading is text", "# Title", `[{"t":"p","c":[{"t":"text","text":"# Title"}]}]`},
		{"table is text", "| a | b |", `[{"t":"p","c":[{"t":"text","text":"| a | b |"}]}]`},
		{
			"line break and paragraphs",
			"a\nb\n\nc",
			`[{"t":"p","c":[{"t":"text","text":"a"},{"t":"br"},{"t":"text","text":"b"}]},{"t":"p","c":[{"t":"text","text":"c"}]}]`,
		},
		{
			"unordered list",
			"- a\n* **b**",
			`[{"t":"ul","c":[{"t":"li","c":[{"t":"text","text":"a"}]},{"t":"li","c":[{"t":"strong","c":[{"t":"text","text":"b"}]}]}]}]`,
		},
		{
			"ordered list", "1. a\n2. b",
			`[{"t":"ol","c":[{"t":"li","c":[{"t":"text","text":"a"}]},{"t":"li","c":[{"t":"text","text":"b"}]}]}]`,
		},
		{
			"blockquote",
			"> quoted\n> more",
			`[{"t":"blockquote","c":[{"t":"p","c":[{"t":"text","text":"quoted"},{"t":"br"},{"t":"text","text":"more"}]}]}]`,
		},
		{
			"unclosed markers stay text", "**open *and `tick",
			`[{"t":"p","c":[{"t":"text","text":"**open *and ` + "`" + `tick"}]}]`,
		},
		{"crlf", "a\r\nb", `[{"t":"p","c":[{"t":"text","text":"a"},{"t":"br"},{"t":"text","text":"b"}]}]`},
		{"empty", "  \n ", `[]`},
	}
	for _, tc := range cases {
		if got := render(t, tc.src); got != tc.want {
			t.Errorf("%s:\n got %s\nwant %s", tc.name, got, tc.want)
		}
	}
}

func TestMarkdownDepthIsBounded(t *testing.T) {
	src := strings.Repeat("> ", 50) + strings.Repeat("**", 200) + "x"
	if got := render(t, src); !strings.Contains(got, "x") {
		t.Fatalf("deep input lost its text: %s", got)
	}
}
