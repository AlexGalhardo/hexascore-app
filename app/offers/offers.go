// Package offers builds the "where to watch, where to buy" section of the item page and resolves the outbound links
// behind it (RFC-003 sections 3 and 4). Rows come from the stages of internal/ingest/offers; this package only reads
// them, adds the affiliate tag at the moment of the redirect (never stored) and counts the click without a user id or
// an IP.
package offers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/uuid"
)

// Countries are the countries the page can switch between, in the order of the switch.
var Countries = []string{"BR", "US", "PT", "ES", "MX", "AR"}

// Everywhere is the country of an offer that is valid in every country (a digital store, an admin link).
const Everywhere = "ZZ"

// kindOrder is the order of the groups on the page: where it is included in a subscription or free first.
var kindOrder = []string{"stream", "free", "rent", "buy", "buy_digital", "mobile", "physical", "read"}

// ErrNotFound is returned for an offer that does not exist, is deactivated or has expired.
var ErrNotFound = errors.New("offers: not found")

// Querier is the read side of the database.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Executor adds writes.
type Executor interface {
	Querier
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// View is one link of the section. The target URL never leaves the server: the page links to /go/{ID}.
type View struct {
	ID    string `json:"id"`
	Store string `json:"store"`
	// Name is the store, for example "Netflix" or "Amazon Brasil".
	Name string `json:"name"`
	// Kind is the offers.kind: stream, free, rent, buy, buy_digital, mobile, physical or read.
	Kind string `json:"kind"`
	// VerifiedAt is the day (UTC, YYYY-MM-DD) the offer was last confirmed, empty when never.
	VerifiedAt string `json:"verifiedAt,omitempty"`
	// Sponsored is true when the store has an affiliate programme and a tag is configured for it.
	Sponsored bool `json:"sponsored"`
}

// Section is everything the item page needs. ByCountry holds, for each switchable country, its offers with the
// everywhere offers merged in, so the HTML is the same for every visitor and the switch needs no request.
type Section struct {
	Countries []string          `json:"countries"`
	Default   string            `json:"defaultCountry"`
	ByCountry map[string][]View `json:"byCountry"`
	// Commission asks for the line "HexaScore may earn a commission": at least one link carries a tag.
	Commission bool `json:"commission"`
	// AmazonAssociate asks for the sentence the Amazon Associates agreement requires.
	AmazonAssociate bool `json:"amazonAssociate"`
	// Attribution lists the data credits owed to the sources shown (for example JustWatch).
	Attribution []string `json:"attribution"`
}

// DefaultCountry is the country a locale starts on: pt-BR is Brazil, es is Spain, everything else the United States.
func DefaultCountry(locale string) string {
	switch locale {
	case "pt-BR":
		return "BR"
	case "es":
		return "ES"
	default:
		return "US"
	}
}

// Tagger returns the affiliate tag for a store code, or "" when the owner has none.
type Tagger func(store string) string

// safeTag is what a tag may look like; anything else would let an environment variable add parameters to a URL.
var safeTag = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// EnvTagger reads AFFILIATE_<STORE>_TAG (amazon_br is AFFILIATE_AMAZON_BR_TAG). An unsafe value counts as unset.
func EnvTagger(get func(string) string) Tagger {
	return func(store string) string {
		tag := get("AFFILIATE_" + strings.ToUpper(store) + "_TAG")
		if !safeTag.MatchString(tag) {
			return ""
		}

		return tag
	}
}

// WithTag adds (or replaces) the affiliate parameter of a URL. Without a parameter name, a tag or a parseable URL it
// returns the URL unchanged.
func WithTag(rawURL, param, tag string) string {
	if param == "" || tag == "" {
		return rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}

	q := u.Query()
	q.Set(param, tag)
	u.RawQuery = q.Encode()

	return u.String()
}

const loadSQL = `SELECT DISTINCT ON (o.country, o.store, o.kind)
  o.id::text, o.country::text, o.store, s.name, s.kind, coalesce(s.affiliate_program, ''), coalesce(s.attribution, ''),
  o.kind, o.last_verified_at, o.source
FROM offers o
JOIN stores s ON s.code = o.store AND s.active
JOIN items i ON i.id = o.item_id AND i.visibility = 'public'
WHERE o.item_id = $1 AND o.active AND (o.expires_at IS NULL OR o.expires_at > $2) AND o.verify_status <> 'gone'
ORDER BY o.country, o.store, o.kind, (o.edition_id IS NOT DISTINCT FROM i.primary_edition_id) DESC, o.id`

// Load builds the section of one item as of now. Per country, store and kind it keeps one offer, the one of the
// primary edition for a book; offers that are deactivated, expired or confirmed gone are left out.
func Load(ctx context.Context, q Querier, itemID string, now time.Time, tag Tagger, locale string) (Section, error) {
	rows, err := q.QueryContext(ctx, loadSQL, itemID, now)
	if err != nil {
		return Section{}, fmt.Errorf("offers load %s: %w", itemID, err)
	}

	defer func() { _ = rows.Close() }()

	sec := Section{
		Countries: slices.Clone(Countries), Default: DefaultCountry(locale),
		ByCountry: make(map[string][]View, len(Countries)), Attribution: []string{},
	}

	byCountry := map[string][]View{}
	amazon := map[string]bool{}

	for rows.Next() {
		var (
			v                                                 View
			country, store, storeKind, programme, attribution string
			offerKind, source                                 string
			verified                                          sql.NullTime
		)

		if err := rows.Scan(
			&v.ID,
			&country,
			&store,
			&v.Name,
			&storeKind,
			&programme,
			&attribution,
			&offerKind,
			&verified,
			&source,
		); err != nil {
			return Section{}, err
		}

		// A watch provider with no known site of its own has nothing to link to: it is left out of the page.
		if source == streamingSource {
			if _, ok := PlatformURL(v.Name); !ok {
				continue
			}
		}

		v.Kind = offerKind
		v.Store = store
		if verified.Valid {
			v.VerifiedAt = verified.Time.UTC().Format(time.DateOnly)
		}

		v.Sponsored = programme != "" && tag(store) != ""
		if v.Sponsored && programme == "amazon_associates" {
			amazon[v.ID] = true
		}

		if attribution != "" && !slices.Contains(sec.Attribution, attribution) {
			sec.Attribution = append(sec.Attribution, attribution)
		}

		byCountry[country] = append(byCountry[country], v)
	}

	if err := rows.Err(); err != nil {
		return Section{}, err
	}

	for _, c := range Countries {
		list := make([]View, 0, len(byCountry[c])+len(byCountry[Everywhere])) // never nil: the page reads a list
		list = append(list, byCountry[c]...)
		list = append(list, byCountry[Everywhere]...)

		slices.SortStableFunc(list, func(a, b View) int {
			if d := slices.Index(kindOrder, a.Kind) - slices.Index(kindOrder, b.Kind); d != 0 {
				return d
			}

			return strings.Compare(a.Name, b.Name)
		})

		sec.ByCountry[c] = list

		for _, v := range list {
			sec.Commission = sec.Commission || v.Sponsored
			sec.AmazonAssociate = sec.AmazonAssociate || amazon[v.ID]
		}
	}

	slices.Sort(sec.Attribution)

	return sec, nil
}

// streamingSource is the offers.source of TMDB watch providers, whose stored link is TMDB's own watch page.
const streamingSource = "tmdb_providers"

// Link is the target behind an offer id.
type Link struct {
	URL     string
	Store   string
	Country string
	// Param is the store's affiliate parameter name, empty when it has no programme.
	Param string
}

// Final is the URL to redirect to: the stored one with the affiliate tag added where a tag exists.
func (l Link) Final(tag Tagger) string { return WithTag(l.URL, l.Param, tag(l.Store)) }

// Resolve finds the active, unexpired offer with this id. Anything that is not a UUID of such an offer is
// ErrNotFound, so the redirect can only ever follow a URL the catalogue stores.
func Resolve(ctx context.Context, q Querier, id string, now time.Time) (Link, error) {
	if _, err := uuid.Parse(id); err != nil {
		return Link{}, ErrNotFound
	}

	rows, err := q.QueryContext(
		ctx,
		`SELECT o.url, o.store, o.country::text, coalesce(s.affiliate_param, ''), o.source, s.name
FROM offers o JOIN stores s ON s.code = o.store AND s.active
JOIN items i ON i.id = o.item_id AND i.visibility = 'public'
WHERE o.id = $1 AND o.active AND (o.expires_at IS NULL OR o.expires_at > $2) AND o.verify_status <> 'gone'`,
		id,
		now,
	)
	if err != nil {
		return Link{}, fmt.Errorf("offers resolve: %w", err)
	}

	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		return Link{}, errors.Join(ErrNotFound, rows.Err())
	}

	var (
		l            Link
		source, name string
	)
	if err := rows.Scan(&l.URL, &l.Store, &l.Country, &l.Param, &source, &name); err != nil {
		return Link{}, err
	}

	// A watch provider links to the service itself, never to TMDB's page about the title.
	if source == streamingSource {
		site, ok := PlatformURL(name)
		if !ok {
			return Link{}, ErrNotFound
		}

		l.URL = site
	}

	return l, nil
}

// Click counts one outbound click for the day, offer and country. There is no user id and no IP in the row.
func Click(ctx context.Context, ex Executor, offerID, country string, now time.Time) error {
	_, err := ex.ExecContext(
		ctx,
		`INSERT INTO offer_clicks (day, offer_id, country, clicks) VALUES ($1::date, $2, $3, 1)
ON CONFLICT (day, offer_id, country) DO UPDATE SET clicks = offer_clicks.clicks + 1`,
		now.UTC(),
		offerID,
		country,
	)
	if err != nil {
		return fmt.Errorf("offers click: %w", err)
	}

	return nil
}
