package webhooks

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/netip"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/crypt"
	vwebhook "github.com/AlexGalhardo/vukem/webhook"
)

// PingResult is one testWebhook attempt. Delivered is true only for a 2xx.
// Status is the receiver's HTTP status, or 0 when the attempt never got one.
type PingResult struct {
	Delivered bool
	Status    int
}

// Ping sends one signed item.batch to an endpoint the user owns. A disabled
// endpoint is ErrNotFound and is not called. KindTest may reach an httptest
// server on 127.0.0.1; production stays on OpenApp (https, no private allow).
func Ping(ctx context.Context, app *vukem.App, userID int64, endpointID string) (PingResult, error) {
	svc, err := openPing(app)
	if err != nil {
		return PingResult{}, err
	}

	var disabled sql.NullTime
	err = svc.db.QueryRowContext(ctx, `SELECT disabled_at FROM webhook_endpoints
WHERE id = $1 AND user_id = $2`, endpointID, userID).Scan(&disabled)
	if errors.Is(err, sql.ErrNoRows) || disabled.Valid {
		return PingResult{}, ErrNotFound
	}

	if err != nil {
		return PingResult{}, err
	}

	now := svc.now().UTC().Format(time.RFC3339)
	body, err := json.Marshal(batchBody{
		Type: "item.batch", DeliveredAt: now,
		Events: []batchEvent{{
			ID: "chg_test", Type: "item.updated", ItemType: "games",
			ItemID: "00000000-0000-0000-0000-000000000000", Fields: []string{"score"}, ChangedAt: now,
		}},
	})
	if err != nil {
		return PingResult{}, err
	}

	id, err := newDeliveryID()
	if err != nil {
		return PingResult{}, err
	}

	_ = svc.deliver(ctx, vwebhook.Delivery{ID: id, Endpoint: endpointID, Payload: body})

	status := svc.attemptStatus(ctx, endpointID, id)

	return PingResult{Delivered: status >= 200 && status < 300, Status: status}, nil
}

// openPing is OpenApp, except in the test kind, where the acceptance server
// is httptest on loopback. Fan-out keeps calling OpenApp.
func openPing(app *vukem.App) (*Service, error) {
	if app == nil || app.Kind != vukem.KindTest {
		return OpenApp(app)
	}

	db, err := app.DB()
	if err != nil {
		return nil, err
	}

	mailer, err := app.Mailer()
	if err != nil {
		return nil, err
	}

	if app.Crypt == nil {
		return nil, crypt.ErrNoKey
	}

	return Open(db, app.Crypt, mailer, app.Now, app.Log, Options{
		AllowHTTP: true,
		Allow:     []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
		Retries:   -1,
	})
}
