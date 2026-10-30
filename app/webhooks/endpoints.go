package webhooks

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/crypt"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/httpclient"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/webhook"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// SecretAD binds webhook signing secrets to this column. Open must use the same bytes.
var SecretAD = []byte("webhook_endpoints.secret_enc")

const maxURLLen = 2048

// Sentinel errors the HTTP layer maps to problem responses.
var (
	ErrPlanRequired = errors.New("webhooks: plan required")
	ErrLimit        = errors.New("webhooks: endpoint limit")
	ErrURL          = errors.New("webhooks: url refused")
	ErrInvalid      = errors.New("webhooks: invalid")
	ErrNotFound     = errors.New("webhooks: not found")
)

// Resolver looks up a hostname. Nil in PublicURL uses the default resolver.
type Resolver func(ctx context.Context, network, host string) ([]netip.Addr, error)

// Draft is a createWebhook body after the typed handler has bound it.
type Draft struct {
	URL       string
	Events    []string
	ItemTypes []string
}

// Endpoint is one registered webhook. Secret is set only by Create.
type Endpoint struct {
	ID        string
	URL       string
	Events    []string
	ItemTypes []string
	CreatedAt time.Time
	Disabled  bool
	Secret    string
}

// PublicURL accepts https URLs whose host resolves only to public addresses.
// A lookup failure is refused: validation fails closed. Delivery still goes
// through httpclient.Safe; this check is the 422 the client sees at create time.
func PublicURL(ctx context.Context, raw string, resolve Resolver) error {
	if len(raw) == 0 || len(raw) > maxURLLen {
		return ErrURL
	}

	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return ErrURL
	}

	host := u.Hostname()
	if host == "" {
		return ErrURL
	}

	if ip, err := netip.ParseAddr(host); err == nil {
		if !httpclient.Allowed(ip, nil) {
			return ErrURL
		}

		return nil
	}

	if resolve == nil {
		resolve = func(ctx context.Context, network, host string) ([]netip.Addr, error) {
			return net.DefaultResolver.LookupNetIP(ctx, network, host)
		}
	}

	ips, err := resolve(ctx, "ip", host)
	if err != nil || len(ips) == 0 {
		return ErrURL
	}

	for _, ip := range ips {
		if !httpclient.Allowed(ip, nil) {
			return ErrURL
		}
	}

	return nil
}

// Create stores one endpoint for a pro or business user. The plaintext secret
// is returned once and only the app.Crypt ciphertext is written.
func Create(ctx context.Context, db *database.DB, enc *crypt.Encrypter, userID int64, plan string, in Draft) (Endpoint, error) {
	cap, ok := planCap(plan)
	if !ok {
		return Endpoint{}, ErrPlanRequired
	}

	events, err := oneOf(in.Events, allowedEvents, 3)
	if err != nil {
		return Endpoint{}, err
	}

	types, err := oneOf(in.ItemTypes, allowedTypes, 4)
	if err != nil {
		return Endpoint{}, err
	}

	storedURL, err := httpsURL(in.URL)
	if err != nil {
		return Endpoint{}, err
	}

	if err := PublicURL(ctx, storedURL, nil); err != nil {
		return Endpoint{}, err
	}

	if enc == nil {
		return Endpoint{}, crypt.ErrNoKey
	}

	secret, err := webhook.NewSecret()
	if err != nil {
		return Endpoint{}, err
	}

	row := Endpoint{
		ID:        uuid.NewV7().String(),
		URL:       storedURL,
		Events:    events,
		ItemTypes: types,
		Secret:    secret,
	}

	err = db.Tx(ctx, func(tx *database.Tx) error {
		var locked int64
		err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&locked)
		if err != nil {
			return err
		}

		var n int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM webhook_endpoints WHERE user_id = $1`, userID).Scan(&n); err != nil {
			return err
		}

		if n >= cap {
			return ErrLimit
		}

		if err := tx.QueryRowContext(ctx, `INSERT INTO webhook_endpoints (id, user_id, url, events, item_types, secret_enc)
VALUES ($1, $2, $3, $4::text[], $5::text[], $6)
RETURNING created_at`,
			row.ID, userID, row.URL, models.TextArray(events), models.TextArray(types), enc.Seal([]byte(secret), SecretAD),
		).Scan(&row.CreatedAt); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return Endpoint{}, err
	}

	return row, nil
}

// List returns the caller's endpoints. It never reads secret_enc.
func List(ctx context.Context, db *database.DB, userID int64) ([]Endpoint, error) {
	rows, err := db.QueryContext(ctx, `SELECT id::text, url, events, item_types, created_at, disabled_at IS NOT NULL
FROM webhook_endpoints WHERE user_id = $1 ORDER BY created_at, id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Endpoint{}
	for rows.Next() {
		var row Endpoint
		var events, types models.TextArray
		if err := rows.Scan(&row.ID, &row.URL, &events, &types, &row.CreatedAt, &row.Disabled); err != nil {
			return nil, err
		}

		row.Events = []string(events)
		row.ItemTypes = []string(types)
		out = append(out, row)
	}

	return out, rows.Err()
}

// Delete removes one endpoint owned by userID. Another user's id is ErrNotFound.
func Delete(ctx context.Context, db *database.DB, userID int64, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrNotFound
	}

	res, err := db.ExecContext(ctx, `DELETE FROM webhook_endpoints WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return ErrNotFound
	}

	return nil
}

// httpsURL folds a case-variant scheme to https://. The column CHECK is
// LIKE 'https://%', which is case-sensitive, while url.Parse accepts any case.
// Storing the raw scheme would fail the insert as an unhandled 500.
func httpsURL(raw string) (string, error) {
	const prefix = "https://"
	if strings.HasPrefix(raw, prefix) {
		return raw, nil
	}

	if len(raw) >= len(prefix) && strings.EqualFold(raw[:len(prefix)], prefix) {
		return prefix + raw[len(prefix):], nil
	}

	return "", ErrURL
}

func planCap(plan string) (int, bool) {
	switch plan {
	case "pro":
		return 2, true
	case "business":
		return 10, true
	default:
		return 0, false
	}
}

var allowedEvents = map[string]struct{}{
	"item.created": {},
	"item.updated": {},
	"item.deleted": {},
}

var allowedTypes = map[string]struct{}{
	"games": {}, "movies": {}, "series": {}, "books": {},
}

func oneOf(values []string, allowed map[string]struct{}, max int) ([]string, error) {
	if len(values) == 0 || len(values) > max {
		return nil, ErrInvalid
	}

	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, raw := range values {
		v := strings.TrimSpace(raw)
		if _, ok := allowed[v]; !ok {
			return nil, ErrInvalid
		}

		if _, dup := seen[v]; dup {
			continue
		}

		seen[v] = struct{}{}
		out = append(out, v)
	}

	if len(out) == 0 {
		return nil, ErrInvalid
	}

	return out, nil
}
