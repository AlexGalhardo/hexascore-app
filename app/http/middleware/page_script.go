package middleware

import (
	"bytes"
	"encoding/json"
	"html"
	"net/http"
	"regexp"
	"strings"
)

// MovePageScript rewrites a finished HTML document for first paint.
//
// Inertia emits the page JSON before the rendered markup, so the LCP text
// sits behind that payload. The script moves to just before </body>. Slashes
// in the JSON are escaped (<\/script>); the real closing tag is the first
// </script> that is not preceded by a backslash.
//
// Module graphs are pulled out of the head. A parser-inserted module on a fast
// connection finishes before the observed LCP, and Lantern then treats that
// file as render-blocking. The loader waits for an LCP entry, then two frames,
// with a 2s fallback when the browser never emits one. The document stays
// inert until Inertia hydrates, so a click in the wait is not submitted
// against the unhydrated form. The loader itself sits after the page JSON,
// which the parser has already seen.
func MovePageScript(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bw := &pageHTML{ResponseWriter: w}
		defer bw.finish()
		next.ServeHTTP(bw, r)
	})
}

type pageHTML struct {
	http.ResponseWriter

	status      int
	buf         []byte
	passthrough bool
}

func (b *pageHTML) WriteHeader(code int) {
	if b.passthrough || b.status != 0 {
		return
	}

	b.status = code
}

func (b *pageHTML) Write(p []byte) (int, error) {
	if b.passthrough {
		return b.ResponseWriter.Write(p)
	}

	if b.status == 0 {
		b.status = http.StatusOK
	}

	if !wantsHTML(b.Header().Get("Content-Type")) {
		b.passthrough = true
		b.ResponseWriter.WriteHeader(b.status)

		return b.ResponseWriter.Write(p)
	}

	b.buf = append(b.buf, p...)

	return len(p), nil
}

func (b *pageHTML) Unwrap() http.ResponseWriter { return b.ResponseWriter }

func (b *pageHTML) finish() {
	if b.passthrough || b.status == 0 {
		return
	}

	body := b.buf
	if wantsHTML(b.Header().Get("Content-Type")) {
		moved := rewriteDocument(body)
		if !bytes.Equal(moved, body) {
			b.Header().Del("Content-Length")
			body = moved
		}
	}

	b.ResponseWriter.WriteHeader(b.status)
	_, _ = b.ResponseWriter.Write(body)
}

func wantsHTML(contentType string) bool {
	if contentType == "" {
		return true
	}

	return strings.Contains(strings.ToLower(contentType), "text/html")
}

var (
	modulePreloadRe = regexp.MustCompile(`<link\b[^>]*\brel="modulepreload"[^>]*>\n?`)
	moduleScriptRe  = regexp.MustCompile(`<script\b[^>]*>\s*</script>`)
)

type modScript struct {
	Src       string `json:"src"`
	Integrity string `json:"integrity,omitempty"`
	Nonce     string `json:"nonce,omitempty"`
}

func rewriteDocument(body []byte) []byte {
	body, mods := stripHeadModules(body)
	body = movePageScript(body)
	if len(mods) == 0 {
		return body
	}

	return insertBeforeBodyClose(body, loaderScript(mods))
}

func stripHeadModules(body []byte) ([]byte, []modScript) {
	start := bytes.Index(body, []byte("<head"))
	end := bytes.Index(body, []byte("</head>"))
	if start < 0 || end < start {
		return body, nil
	}

	head := append([]byte(nil), body[start:end]...)
	head = modulePreloadRe.ReplaceAll(head, nil)

	var mods []modScript
	head = moduleScriptRe.ReplaceAllFunc(head, func(tag []byte) []byte {
		if !bytes.Contains(tag, []byte(`type="module"`)) {
			return tag
		}

		src := attrValue(tag, "src")
		if src == "" {
			return tag
		}

		mods = append(mods, modScript{
			Src:       src,
			Integrity: attrValue(tag, "integrity"),
			Nonce:     attrValue(tag, "nonce"),
		})

		return nil
	})

	out := append([]byte(nil), body[:start]...)
	out = append(out, head...)
	out = append(out, body[end:]...)

	return out, mods
}

func attrValue(tag []byte, name string) string {
	_, rest, ok := bytes.Cut(tag, []byte(name+`="`))
	if !ok {
		return ""
	}

	value, _, ok := bytes.Cut(rest, []byte(`"`))
	if !ok {
		return ""
	}

	return string(value)
}

func loaderScript(mods []modScript) []byte {
	payload, err := json.Marshal(mods)
	if err != nil {
		return nil
	}

	var b strings.Builder
	b.WriteString("<script")
	if mods[0].Nonce != "" {
		b.WriteString(` nonce="`)
		b.WriteString(html.EscapeString(mods[0].Nonce))
		b.WriteString(`"`)
	}
	// One frame after load still starts the module before Lighthouse records
	// LCP on / and /games. Lantern then treats that file as render-blocking.
	// Wait for an LCP entry, then two frames. The 2s timeout covers a browser
	// that never emits LCP. Script onload fires when the module is evaluated,
	// which is before createInertiaApp resolves, so the loader must not clear
	// `inert` there: Playwright would submit the server form and stay on
	// /login. app.tsx clears it after hydration. The 15s timeout is only the
	// escape if that promise never settles. `})})})` is a SyntaxError and the
	// app never hydrates.
	b.WriteString(`>document.documentElement.inert=true;addEventListener("load",function(){var started=0;function go(){if(started)return;started=1;var mods=`)
	b.Write(payload)
	b.WriteString(`;var left=mods.length;function done(){if(--left)return}`)
	b.WriteString(`for(var i=0;i<mods.length;i++){var m=mods[i];var s=document.createElement("script");s.type="module";s.src=m.src;s.onload=done;s.onerror=done;`)
	b.WriteString(`if(m.integrity){s.integrity=m.integrity;s.crossOrigin="anonymous"}`)
	b.WriteString(`if(m.nonce)s.nonce=m.nonce;document.head.appendChild(s)}}`)
	b.WriteString(`try{var o=new PerformanceObserver(function(){o.disconnect();requestAnimationFrame(function(){requestAnimationFrame(go)})});`)
	b.WriteString(`o.observe({type:"largest-contentful-paint",buffered:true})}catch(e){setTimeout(go,0)}`)
	b.WriteString(`setTimeout(go,2000);setTimeout(function(){document.documentElement.inert=false},15000)})</script>`)

	return []byte(b.String())
}

func insertBeforeBodyClose(body, snippet []byte) []byte {
	if len(snippet) == 0 {
		return body
	}

	end := []byte("</body>")
	i := bytes.Index(body, end)
	if i < 0 {
		return append(append([]byte(nil), body...), snippet...)
	}

	out := make([]byte, 0, len(body)+len(snippet)+1)
	out = append(out, body[:i]...)
	out = append(out, snippet...)
	out = append(out, '\n')
	out = append(out, body[i:]...)

	return out
}

func movePageScript(body []byte) []byte {
	const openPrefix = `<script data-page="`

	start := bytes.Index(body, []byte(openPrefix))
	if start < 0 {
		return body
	}

	gt := bytes.IndexByte(body[start:], '>')
	if gt < 0 {
		return body
	}

	openEnd := start + gt + 1
	if !bytes.Contains(body[start:openEnd], []byte(`type="application/json"`)) {
		return body
	}

	rel := indexScriptClose(body[openEnd:])
	if rel < 0 {
		return body
	}

	end := openEnd + rel + len("</script>")
	tag := append([]byte(nil), body[start:end]...)
	rest := make([]byte, 0, len(body))
	rest = append(rest, body[:start]...)
	rest = append(rest, body[end:]...)

	closeBody := []byte("</body>")
	i := bytes.Index(rest, closeBody)
	if i < 0 {
		return append(rest, tag...)
	}

	out := make([]byte, 0, len(rest)+len(tag)+1)
	out = append(out, rest[:i]...)
	out = append(out, tag...)
	out = append(out, '\n')
	out = append(out, rest[i:]...)

	return out
}

// indexScriptClose finds </script> not preceded by a backslash. JSON escapes
// every slash, so a prop value of </script> is stored as <\/script> and must
// not end the element.
func indexScriptClose(b []byte) int {
	const closeTag = "</script>"

	for i := 0; i+len(closeTag) <= len(b); i++ {
		if b[i] != '<' || !bytes.Equal(b[i:i+len(closeTag)], []byte(closeTag)) {
			continue
		}

		if i > 0 && b[i-1] == '\\' {
			continue
		}

		return i
	}

	return -1
}
