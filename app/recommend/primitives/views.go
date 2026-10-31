package primitives

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AlexGalhardo/vukem/cache"
)

// MaxItemsPerVisitor is the number of distinct items one visitor can add views to per UTC day (SDD 08 section 7.1).
const MaxItemsPerVisitor = 50

// dedupEntries bounds the in-memory dedup store (SDD 08 section 7.1).
const dedupEntries = 300_000

// Recorder counts item views and search clicks. Build one with [For] (or in a test, directly).
type Recorder struct {
	counter *Counter
	seen    *cache.Cache
	salt    func(day string) ([]byte, error)
	now     func() time.Time
	db      dbFunc

	mu      sync.Mutex
	visited map[string]int // HMAC(salt, visitor) -> distinct items counted today
	visDay  string
}

// CountView counts one view of the item at (typ, ref) for visitor unless that visitor already viewed it today (UTC)
// or already counted MaxItemsPerVisitor items. The dedup key is HMAC-SHA256(daily salt, visitor|item): the visitor
// (a session or an IP) is never stored, and the salt changes at 00:00 UTC.
func (r *Recorder) CountView(ctx context.Context, visitor, typ, ref string) bool {
	now := r.now().UTC()
	day := now.Format(time.DateOnly)

	salt, err := r.salt(day)
	if err != nil {
		return false
	}

	claim := mac(salt, visitor+"|"+typ+"/"+ref)
	// The memory store never errors; on one, skipping the view is safer than counting it twice.
	fresh, err := cache.Add(ctx, r.seen, claim, true, time.Until(midnightAfter(now)))
	if err != nil || !fresh {
		return false
	}

	if !r.underCap(day, mac(salt, visitor)) {
		return false
	}

	r.counter.Add(Key{Type: typ, Ref: ref, Day: day}, 1, 0)

	return true
}

// CountClick counts one search click on the item with id itemID.
func (r *Recorder) CountClick(itemID string, at time.Time) {
	r.counter.Add(Key{Ref: itemID, Day: at.UTC().Format(time.DateOnly)}, 0, 1)
}

func (r *Recorder) underCap(day, visitor string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.visDay != day || len(r.visited) >= dedupEntries {
		r.visited, r.visDay = map[string]int{}, day
	}

	if r.visited[visitor] >= MaxItemsPerVisitor {
		return false
	}

	r.visited[visitor]++

	return true
}

func mac(salt []byte, s string) string {
	h := hmac.New(sha256.New, salt)
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

func midnightAfter(t time.Time) time.Time {
	y, m, d := t.Date()

	return time.Date(y, m, d+1, 0, 0, 0, 0, time.UTC)
}

var botTokens = []string{
	"bot", "crawl", "spider", "slurp", "headless", "lighthouse", "facebookexternalhit", "preview",
	"curl/", "wget/", "python-requests", "go-http-client",
}

// Countable reports whether a request may add a view: not a prefetch, not an Inertia partial reload and not a bot.
func Countable(r *http.Request) bool {
	if strings.Contains(strings.ToLower(r.Header.Get("Sec-Purpose")), "prefetch") ||
		strings.EqualFold(r.Header.Get("Purpose"), "prefetch") ||
		strings.EqualFold(r.Header.Get("X-Moz"), "prefetch") {
		return false
	}

	if r.Header.Get("X-Inertia-Partial-Data") != "" || r.Header.Get("X-Inertia-Partial-Component") != "" {
		return false
	}

	ua := strings.ToLower(r.Header.Get("User-Agent"))
	for _, t := range botTokens {
		if strings.Contains(ua, t) {
			return false
		}
	}

	return true
}

var sectionTypes = map[string]string{"games": "game", "movies": "movie", "series": "series", "books": "book"}

// Target maps a request path to the item it reads: /{section}/{slug} (after an optional locale prefix) for the item
// pages and /app/v1/{section}/{id} for the app reads. A zero Key means the path is not an item read.
func Target(path string, localePrefixes map[string]string) (k Key, app bool) {
	segs := strings.Split(strings.Trim(path, "/"), "/")

	if _, ok := localePrefixes[strings.ToLower(segs[0])]; ok {
		segs = segs[1:]
	}

	if len(segs) >= 2 && segs[0] == "app" && segs[1] == "v1" {
		app, segs = true, segs[2:]
	}

	if len(segs) != 2 || segs[1] == "" || (app && segs[1] == "genres") {
		return Key{}, false
	}

	typ, ok := sectionTypes[segs[0]]
	if !ok {
		return Key{}, false
	}

	return Key{Type: typ, Ref: segs[1]}, app
}
