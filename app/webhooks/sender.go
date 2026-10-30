package webhooks

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/crypt"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/mail"
	"github.com/AlexGalhardo/vukem/queue"
	vwebhook "github.com/AlexGalhardo/vukem/webhook"

	"github.com/AlexGalhardo/hexascore-app/app/metrics"
)

// DeliveryRetries is how many times a failed delivery is retried (SDD 05 §7.3:
// 8 from 30s). Zero on Options means this value. A negative Options.Retries
// means the first attempt is the last one, which tests use.
const DeliveryRetries = 8

const failureLimit = 3

// Options tune a Service. The zero value is the production sender: https only,
// no private addresses, DeliveryRetries, backoff from 30s.
type Options struct {
	AllowHTTP bool
	Allow     []netip.Prefix
	Retries   int
	Now       func() time.Time
}

// Service delivers signed batches and counts exhausted failures.
type Service struct {
	db      *database.DB
	mailer  *mail.Mailer
	sender  *vwebhook.Sender
	now     func() time.Time
	retries int
	log     *slog.Logger
}

// OpenApp is the production service: the app encrypter, mailer and clock.
func OpenApp(app *vukem.App) (*Service, error) {
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

	return Open(db, app.Crypt, mailer, app.Now, app.Log, Options{})
}

// Open builds a service. enc opens secret_enc. mailer may be nil in a test
// that does not exhaust an endpoint. log may be nil.
func Open(db *database.DB, enc *crypt.Encrypter, mailer *mail.Mailer, now func() time.Time, log *slog.Logger, opt Options) (*Service, error) {
	if now == nil {
		now = time.Now
	}

	if log == nil {
		log = slog.Default()
	}

	attempts, err := vwebhook.NewSQLDeliveryLog(db, "")
	if err != nil {
		return nil, err
	}

	retries := normalizeRetries(opt.Retries)
	sender, err := vwebhook.NewSender(vwebhook.SenderOptions{
		Endpoint: func(ctx context.Context, id string) (vwebhook.Endpoint, error) {
			return lookupEndpoint(ctx, db, enc, id)
		},
		Log:               attempts,
		OnGone:            func(ctx context.Context, id string) error { return disableEndpoint(ctx, db, id) },
		Logger:            log,
		Now:               now,
		Retries:           retries,
		Backoff:           queue.Exponential(30 * time.Second),
		InsecureAllowHTTP: opt.AllowHTTP,
		InsecureAllow:     opt.Allow,
	})
	if err != nil {
		return nil, err
	}

	return &Service{db: db, mailer: mailer, sender: sender, now: now, retries: retries, log: log}, nil
}

func normalizeRetries(n int) int {
	switch {
	case n < 0:
		return 0
	case n == 0:
		return DeliveryRetries
	default:
		return n
	}
}

// Registration is the webhook.deliver handler. Retries and the 30s backoff
// live on the queue registration; Sender.Deliver performs one attempt.
func (s *Service) Registration() queue.Registration {
	return queue.HandleKind(vwebhook.DeliverKind, s.deliver,
		queue.Retries(s.retries), queue.Backoff(queue.Exponential(30*time.Second)), queue.Timeout(20*time.Second))
}

// DeliveryRegistration is Registration built from the app on the first delivery.
func DeliveryRegistration(app *vukem.App) queue.Registration {
	return queue.HandleKind(vwebhook.DeliverKind, func(ctx context.Context, d vwebhook.Delivery) error {
		svc, err := OpenApp(app)
		if err != nil {
			return err
		}

		return svc.deliver(ctx, d)
	}, queue.Retries(DeliveryRetries), queue.Backoff(queue.Exponential(30*time.Second)), queue.Timeout(20*time.Second))
}

func (s *Service) deliver(ctx context.Context, d vwebhook.Delivery) error {
	err := s.sender.Deliver(ctx, d)
	metrics.RecordWebhookDelivery(ctx, s.deliveryOutcome(ctx, d.Endpoint, err))
	if err == nil {
		if resetErr := resetStreak(ctx, s.db, d.Endpoint); resetErr != nil {
			s.log.ErrorContext(ctx, "webhook streak reset", "endpoint", d.Endpoint, "err", resetErr)
		}

		return nil
	}

	if a, ok := queue.AttemptFrom(ctx); ok && a.Last() {
		if noteErr := s.noteExhausted(ctx, d.Endpoint); noteErr != nil {
			s.log.ErrorContext(ctx, "webhook streak", "endpoint", d.Endpoint, "err", noteErr)
		}
	}

	return err
}

// deliveryOutcome is the api.webhook.deliveries label. A non-nil error is a
// failure. A nil error is success, 410 or a disabled endpoint, which the
// delivery log distinguishes.
func (s *Service) deliveryOutcome(ctx context.Context, endpoint string, deliverErr error) string {
	if deliverErr != nil {
		return "failure"
	}

	var status int
	var attemptErr string
	err := s.db.QueryRowContext(ctx, `SELECT status, error FROM webhook_deliveries
WHERE endpoint = $1 ORDER BY id DESC LIMIT 1`, endpoint).Scan(&status, &attemptErr)
	if err != nil {
		return "success"
	}

	switch {
	case status >= 200 && status < 300:
		return "success"
	case status == 410:
		return "gone"
	case attemptErr == vwebhook.AttemptDisabled:
		return "disabled"
	default:
		return "failure"
	}
}

func (s *Service) latestStatus(ctx context.Context, endpoint string) int {
	var status int
	err := s.db.QueryRowContext(ctx, `SELECT status FROM webhook_deliveries
WHERE endpoint = $1 ORDER BY id DESC LIMIT 1`, endpoint).Scan(&status)
	if err != nil {
		return 0
	}

	return status
}

func lookupEndpoint(ctx context.Context, db *database.DB, enc *crypt.Encrypter, id string) (vwebhook.Endpoint, error) {
	var rawURL, stored string
	var disabled sql.NullTime
	err := db.QueryRowContext(ctx, `SELECT url, secret_enc, disabled_at FROM webhook_endpoints WHERE id = $1`, id).
		Scan(&rawURL, &stored, &disabled)
	if errors.Is(err, sql.ErrNoRows) || disabled.Valid {
		return vwebhook.Endpoint{}, vwebhook.ErrEndpointDisabled
	}

	if err != nil {
		return vwebhook.Endpoint{}, err
	}

	plain, err := enc.Open(stored, SecretAD)
	if err != nil {
		return vwebhook.Endpoint{}, fmt.Errorf("webhook: open secret: %w", err)
	}

	return vwebhook.Endpoint{URL: rawURL, Secret: string(plain)}, nil
}

func disableEndpoint(ctx context.Context, db *database.DB, id string) error {
	_, err := db.ExecContext(ctx, `UPDATE webhook_endpoints SET disabled_at = now() WHERE id = $1 AND disabled_at IS NULL`, id)

	return err
}

func resetStreak(ctx context.Context, db *database.DB, id string) error {
	_, err := db.ExecContext(ctx, `UPDATE webhook_endpoints SET failure_streak = 0 WHERE id = $1 AND failure_streak <> 0`, id)

	return err
}

// noteExhausted counts one delivery that will not be retried. The third one
// disables the endpoint and mails its owner once.
func (s *Service) noteExhausted(ctx context.Context, id string) error {
	var streak int
	err := s.db.QueryRowContext(ctx, `UPDATE webhook_endpoints
SET failure_streak = failure_streak + 1
WHERE id = $1 AND disabled_at IS NULL
RETURNING failure_streak`, id).Scan(&streak)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	if err != nil {
		return err
	}

	if streak < failureLimit {
		return nil
	}

	if _, err := s.db.ExecContext(ctx, `UPDATE webhook_endpoints SET disabled_at = now() WHERE id = $1 AND disabled_at IS NULL`, id); err != nil {
		return err
	}

	if streak != failureLimit || s.mailer == nil {
		return nil
	}

	var email string
	err = s.db.QueryRowContext(ctx, `SELECT u.email FROM webhook_endpoints e JOIN users u ON u.id = e.user_id WHERE e.id = $1`, id).Scan(&email)
	if err != nil || email == "" {
		return err
	}

	if err := s.mailer.Send(ctx, disabledNotice{To: email, EndpointID: id}); err != nil {
		s.log.ErrorContext(ctx, "webhook disabled mail", "endpoint", id, "err", err)
	}

	return nil
}

type disabledNotice struct {
	To         string
	EndpointID string
}

func (n disabledNotice) Envelope() mail.Envelope {
	return mail.Envelope{To: []string{n.To}, Subject: "Webhook endpoint disabled"}
}

func (n disabledNotice) Content() mail.Content {
	return mail.Text("HexaScore disabled webhook endpoint " + n.EndpointID +
		" after three deliveries in a row failed. Register it again after the receiver is fixed. Missed changes stay on the changes feed.")
}

func newDeliveryID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("webhook: delivery id: %w", err)
	}

	return "msg_" + hex.EncodeToString(b), nil
}
