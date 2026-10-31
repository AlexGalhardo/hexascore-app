// Package v1 serves the webhook management routes on /api/v1 (SDD 05 §6.2).
package v1

import (
	"errors"
	"net/http"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/crypt"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/webhooks"
)

// Webhooks is listWebhooks, createWebhook, deleteWebhook and testWebhook.
type Webhooks struct {
	App   *vukem.App
	DB    *database.DB
	Crypt *crypt.Encrypter
	Plans api.PlanResolver
}

// CreateWebhook is the createWebhook body.
type CreateWebhook struct {
	URL       string   `json:"url"       validate:"required"`
	Events    []string `json:"events"    validate:"required"`
	ItemTypes []string `json:"itemTypes" validate:"required"`
}

// WebhookJSON is one endpoint. The signing secret is never on this shape.
type WebhookJSON struct {
	ID        string   `json:"id"        validate:"required"`
	URL       string   `json:"url"       validate:"required"`
	Events    []string `json:"events"    validate:"required"`
	ItemTypes []string `json:"itemTypes" validate:"required"`
	CreatedAt string   `json:"createdAt" validate:"required"`
	Disabled  bool     `json:"disabled"`
}

// WebhookCreated is the 201 body. Secret is shown once.
type WebhookCreated struct {
	ID        string   `json:"id"        validate:"required"`
	URL       string   `json:"url"       validate:"required"`
	Events    []string `json:"events"    validate:"required"`
	ItemTypes []string `json:"itemTypes" validate:"required"`
	CreatedAt string   `json:"createdAt" validate:"required"`
	Disabled  bool     `json:"disabled"`
	Secret    string   `json:"secret"    validate:"required"`
}

// WebhookList is listWebhooks.
type WebhookList struct {
	Data []WebhookJSON `json:"data"`
}

// WebhookTestResult is testWebhook. Delivered is true only when the receiver
// answered 2xx. Status is that HTTP status, or 0 when there was none.
type WebhookTestResult struct {
	Delivered bool `json:"delivered"`
	Status    int  `json:"status"`
}

// WebhookID is the deleteWebhook path.
type WebhookID struct {
	ID string `json:"id" validate:"required"`
}

// List is GET /webhooks.
func (h *Webhooks) List(c *web.Ctx, _ struct{}) (WebhookList, error) {
	id, err := h.userID(c)
	if err != nil {
		return WebhookList{}, err
	}

	rows, err := webhooks.List(c.Context(), h.DB, id)
	if err != nil {
		return WebhookList{}, err
	}

	out := make([]WebhookJSON, 0, len(rows))
	for _, row := range rows {
		out = append(out, toJSON(row))
	}

	return WebhookList{Data: out}, nil
}

// Create is POST /webhooks.
func (h *Webhooks) Create(c *web.Ctx, in CreateWebhook) (WebhookCreated, error) {
	id, err := h.userID(c)
	if err != nil {
		return WebhookCreated{}, err
	}

	plan, err := h.Plans(c.Context(), id)
	if err != nil {
		return WebhookCreated{}, api.Unavailable(err)
	}

	row, err := webhooks.Create(c.Context(), h.DB, h.Crypt, id, plan.ID, webhooks.Draft{
		URL: in.URL, Events: in.Events, ItemTypes: in.ItemTypes,
	})
	if err != nil {
		return WebhookCreated{}, asProblem(err)
	}

	body := toJSON(row)

	return WebhookCreated{
		ID: body.ID, URL: body.URL, Events: body.Events, ItemTypes: body.ItemTypes,
		CreatedAt: body.CreatedAt, Disabled: body.Disabled, Secret: row.Secret,
	}, nil
}

// Delete is DELETE /webhooks/{id}.
func (h *Webhooks) Delete(c *web.Ctx, in WebhookID) (struct{}, error) {
	id, err := h.userID(c)
	if err != nil {
		return struct{}{}, err
	}

	if err := webhooks.Delete(c.Context(), h.DB, id, in.ID); err != nil {
		return struct{}{}, asProblem(err)
	}

	return struct{}{}, nil
}

// Test is POST /webhooks/{id}/test. It sends one signed sample batch.
func (h *Webhooks) Test(c *web.Ctx, in WebhookID) (WebhookTestResult, error) {
	id, err := h.userID(c)
	if err != nil {
		return WebhookTestResult{}, err
	}

	plan, err := h.Plans(c.Context(), id)
	if err != nil {
		return WebhookTestResult{}, api.Unavailable(err)
	}

	if plan.ID != api.PlanPro && plan.ID != api.PlanBusiness {
		return WebhookTestResult{}, errs.New(http.StatusForbidden, "plan required").
			WithExtension("code", "plan-required")
	}

	got, err := webhooks.Ping(c.Context(), h.App, id, in.ID)
	if err != nil {
		return WebhookTestResult{}, asProblem(err)
	}

	return WebhookTestResult{Delivered: got.Delivered, Status: got.Status}, nil
}

func (h *Webhooks) userID(c *web.Ctx) (int64, error) {
	u, err := auth.User[models.User](c, api.GuardName)
	if err != nil {
		return 0, err
	}

	return u.ID, nil
}

func toJSON(row webhooks.Endpoint) WebhookJSON {
	events := row.Events
	if events == nil {
		events = []string{}
	}

	types := row.ItemTypes
	if types == nil {
		types = []string{}
	}

	return WebhookJSON{
		ID: row.ID, URL: row.URL, Events: events, ItemTypes: types,
		CreatedAt: row.CreatedAt.UTC().Format(time.RFC3339), Disabled: row.Disabled,
	}
}

func asProblem(err error) error {
	switch {
	case errors.Is(err, webhooks.ErrPlanRequired):
		return errs.New(http.StatusForbidden, "plan required").WithExtension("code", "plan-required")
	case errors.Is(err, webhooks.ErrLimit):
		return errs.New(http.StatusUnprocessableEntity, "webhook endpoint limit reached").
			WithExtension("code", "webhook-limit")
	case errors.Is(err, webhooks.ErrURL):
		return errs.New(http.StatusUnprocessableEntity, "webhook url refused").WithExtension("code", "webhook-url")
	case errors.Is(err, webhooks.ErrInvalid):
		return errs.New(http.StatusUnprocessableEntity, "webhook fields refused").
			WithExtension("code", "webhook-invalid")
	case errors.Is(err, webhooks.ErrNotFound):
		return errs.ErrNotFound
	case errors.Is(err, crypt.ErrNoKey):
		return api.Unavailable(err)
	default:
		return err
	}
}
