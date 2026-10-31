package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMovePageScript(t *testing.T) {
	json := `{"component":"Catalog/Index","props":{"intro":"x<\/script>"}}`
	page := `<head>` +
		`<link rel="stylesheet" href="/build/assets/app.css">` +
		`<link rel="modulepreload" href="/build/assets/jsx.js" integrity="sha384-abc" nonce="n">` +
		`<script type="module" src="/build/assets/app.js" integrity="sha384-app" nonce="n"></script>` +
		`</head><body>` +
		`<script data-page="app" type="application/json">` + json + `</script>` +
		`<div id="app"><p class="mt-1">intro</p></div>` +
		`<script type="module" src="/not-head.js"></script>` +
		`</body>`
	h := MovePageScript(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", "999")
		_, _ = w.Write([]byte(page))
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/games", nil))
	body := rec.Body.String()

	if strings.Count(body, `data-page="app"`) != 1 {
		t.Fatalf("script count: %s", body)
	}

	if strings.Index(body, "mt-1") > strings.Index(body, "data-page") {
		t.Fatalf("LCP text is after the page JSON: %s", body)
	}

	if !strings.Contains(body, json+`</script>`) {
		t.Fatalf("page JSON broken: %s", body)
	}

	if strings.Contains(body, "modulepreload") {
		t.Fatalf("modulepreload still in the head: %s", body)
	}

	if strings.Contains(body, `<script type="module" src="/build/assets/app.js"`) {
		t.Fatalf("parser still inserts the module: %s", body)
	}

	if !strings.Contains(body, `href="/build/assets/app.css"`) {
		t.Fatalf("stylesheet dropped: %s", body)
	}

	if !strings.Contains(
		body,
		`document.documentElement.inert=true;addEventListener("load",function(){var started=0;function go(){if(started)return;started=1;var mods=`,
	) ||
		!strings.Contains(body, `/build/assets/app.js`) ||
		!strings.Contains(body, `function done(){if(--left)return}`) ||
		!strings.Contains(body, `s.onload=done;s.onerror=done;`) ||
		!strings.Contains(body, `setTimeout(function(){document.documentElement.inert=false},15000)`) {
		t.Fatalf("loader missing: %s", body)
	}

	if strings.Contains(body, `function done(){if(--left)return;document.documentElement.inert=false}`) {
		t.Fatalf("loader clears inert on script load, before hydration: %s", body)
	}

	if !strings.Contains(
		body,
		`setTimeout(go,2000);setTimeout(function(){document.documentElement.inert=false},15000)})</script>`,
	) {
		t.Fatalf("loader is not valid JavaScript: %s", body)
	}

	if !strings.Contains(body, `nonce="n"`) {
		t.Fatalf("loader lost the nonce: %s", body)
	}

	if !strings.Contains(body, `src="/not-head.js"`) {
		t.Fatalf("body module was rewritten: %s", body)
	}

	loaderAt := strings.Index(body, `addEventListener("load"`)
	pageAt := strings.Index(body, "data-page")
	if loaderAt < pageAt {
		t.Fatalf("loader runs before the page JSON is parsed: %s", body)
	}

	if rec.Header().Get("Content-Length") != "" {
		t.Fatalf("content-length kept after rewrite: %s", rec.Header().Get("Content-Length"))
	}
}

func TestMovePageScript_LeavesCSS(t *testing.T) {
	const css = `<script data-page="app" type="application/json">{}</script>`
	h := MovePageScript(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Content-Length", "54")
		_, _ = w.Write([]byte(css))
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/build/app.css", nil))

	if rec.Body.String() != css {
		t.Fatalf("css body changed: %s", rec.Body.String())
	}

	if rec.Header().Get("Content-Length") != "54" {
		t.Fatalf("content-length = %q", rec.Header().Get("Content-Length"))
	}
}
