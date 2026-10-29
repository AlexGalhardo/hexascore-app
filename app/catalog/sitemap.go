package catalog

import (
	"bufio"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
)

// SitemapChunk is the URL cap of one child sitemap: under the protocol's
// 50,000 and SDD 09's 40,000 (FR-09-11).
const SitemapChunk = 40000

// MinIndexableItems is the linked public items an entity needs to be listed
// and indexable (open issues 01-8, 09-8); below it the page is noindex,follow.
const MinIndexableItems = 3

// ErrNoSitemap means the requested child sitemap does not exist.
var ErrNoSitemap = errors.New("catalog: no such sitemap")

// sitemapSource describes one family of child sitemaps.
type sitemapSource struct {
	family string // "items" or "entities"
	key    string // type or kind
	path   string // web section
	column string // entity column holding the URL key
}

var sitemapSources = []sitemapSource{
	{"items", "game", "games", ""},
	{"items", "movie", "movies", ""},
	{"items", "series", "series", ""},
	{"items", "book", "books", ""},
	{"entities", "person", "people", "slug"},
	{"entities", "organization", "companies", "slug"},
	{"entities", "franchise", "franchises", "slug"},
	{"entities", "platform", "platforms", "slug"},
	{"entities", "genre", "genres", "slug"},
	{"entities", "game_mode", "game-modes", "slug"},
	{"entities", "country", "countries", "code"},
	{"entities", "language", "languages", "code"},
}

var staticPaths = []string{"/", "/games", "/movies", "/series", "/books"}

func (s sitemapSource) name(n int) string { return fmt.Sprintf("%s-%s-%d.xml", s.family, s.key, n) }

// query selects the URL keys of one source; column is a constant, never input.
func (s sitemapSource) query() string {
	if s.family == "items" {
		return `SELECT slug FROM items WHERE type = $1 AND visibility = 'public' ORDER BY id OFFSET $2 LIMIT $3`
	}

	return fmt.Sprintf(`SELECT e.%[1]s FROM entities e
JOIN item_entities ie ON ie.entity_id = e.id
JOIN items i ON i.id = ie.item_id AND i.visibility = 'public'
WHERE e.kind = $1 AND e.%[1]s IS NOT NULL
GROUP BY e.id HAVING count(DISTINCT ie.item_id) >= %[2]d
ORDER BY e.id OFFSET $2 LIMIT $3`, s.column, MinIndexableItems)
}

func (s sitemapSource) count(ctx context.Context, ex database.Executor) (int, error) {
	var n int

	err := ex.QueryRowContext(ctx, "SELECT count(*) FROM ("+strings.Replace(s.query(), "OFFSET $2 LIMIT $3", "", 1)+") q",
		s.key).
		Scan(&n)

	return n, err
}

func esc(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))

	return b.String()
}

// writeURL writes one <url> with the locale alternates of path; x-default is English.
func writeURL(w io.Writer, base, path string) error {
	loc := func(prefix string) string {
		if path == "/" && prefix != "" {
			return base + prefix
		}

		return base + prefix + path
	}

	_, err := fmt.Fprintf(w, `<url><loc>%[1]s</loc>`+
		`<xhtml:link rel="alternate" hreflang="en" href="%[1]s"/>`+
		`<xhtml:link rel="alternate" hreflang="pt-br" href="%[2]s"/>`+
		`<xhtml:link rel="alternate" hreflang="es" href="%[3]s"/>`+
		`<xhtml:link rel="alternate" hreflang="x-default" href="%[1]s"/></url>`+"\n",
		esc(loc("")), esc(loc("/pt-br")), esc(loc("/es")))

	return err
}

const urlsetOpen = `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
	`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">` + "\n"

// WriteSitemapIndex writes the index of the static and non-empty child
// sitemaps. base is the site origin without a trailing slash.
func WriteSitemapIndex(ctx context.Context, ex database.Executor, w io.Writer, base string) error {
	bw := bufio.NewWriter(w)
	_, _ = io.WriteString(bw, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
		`<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`+"\n")
	_, _ = fmt.Fprintf(bw, "<sitemap><loc>%s/sitemaps/static.xml</loc></sitemap>\n", esc(base))

	for _, s := range sitemapSources {
		total, err := s.count(ctx, ex)
		if err != nil {
			return fmt.Errorf("count %s %s: %w", s.family, s.key, err)
		}

		for n := 1; n <= (total+SitemapChunk-1)/SitemapChunk; n++ {
			_, _ = fmt.Fprintf(bw, "<sitemap><loc>%s/sitemaps/%s</loc></sitemap>\n", esc(base), s.name(n))
		}
	}

	_, _ = io.WriteString(bw, "</sitemapindex>\n")

	return bw.Flush()
}

// WriteSitemap streams the child sitemap called name (`static.xml`,
// `items-game-1.xml`, `entities-person-2.xml`), or returns ErrNoSitemap. It
// writes nothing before it knows the file exists.
func WriteSitemap(ctx context.Context, ex database.Executor, w io.Writer, base, name string) error {
	if name == "static.xml" {
		bw := bufio.NewWriter(w)
		_, _ = io.WriteString(bw, urlsetOpen)

		for _, p := range staticPaths {
			if err := writeURL(bw, base, p); err != nil {
				return err
			}
		}

		_, _ = io.WriteString(bw, "</urlset>\n")

		return bw.Flush()
	}

	for _, s := range sitemapSources {
		num, ok := strings.CutPrefix(name, s.family+"-"+s.key+"-")
		num, okSuffix := strings.CutSuffix(num, ".xml")

		n, err := strconv.Atoi(num)
		if ok && okSuffix && err == nil && n >= 1 && s.name(n) == name {
			return writeChunk(ctx, ex, w, base, s, n)
		}
	}

	return ErrNoSitemap
}

func writeChunk(ctx context.Context, ex database.Executor, w io.Writer, base string, s sitemapSource, n int) error {
	rows, err := ex.QueryContext(ctx, s.query(), s.key, (n-1)*SitemapChunk, SitemapChunk)
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()

	bw := bufio.NewWriter(w)
	written := 0

	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return err
		}

		if written == 0 {
			_, _ = io.WriteString(bw, urlsetOpen)
		}

		written++

		if err := writeURL(bw, base, "/"+s.path+"/"+key); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if written == 0 {
		return ErrNoSitemap
	}

	_, _ = io.WriteString(bw, "</urlset>\n")

	return bw.Flush()
}
