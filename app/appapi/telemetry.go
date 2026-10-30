package appapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"
)

const (
	reportMessageMax = 1000
	reportStackMax   = 8192
	reportRouteMax   = 200
	reportVersionMax = 64
)

var vitalNames = map[string]struct{}{
	"LCP": {}, "INP": {}, "CLS": {}, "TTFB": {},
}

// ClientError stores one anonymous crash report and answers 204.
// Unknown JSON keys are 400. The row has no user id, token id or e-mail.
func ClientError(db *database.DB, now func() time.Time) func(*web.Ctx) error {
	return func(c *web.Ctx) error {
		raw, err := decodeObject(c)
		if err != nil {
			return err
		}

		if err := rejectUnknown(raw, "appVersion", "versionCode", "androidApi", "screen", "message", "stack"); err != nil {
			return err
		}

		row, err := errorReport(raw)
		if err != nil {
			return err
		}

		if err := insertReport(c.Request().Context(), db, row, now().UTC()); err != nil {
			return err
		}

		c.Log().Info("client.error", "route", row.route)

		return c.NoContent()
	}
}

// ClientMetric stores one anonymous web-vital beacon and answers 204.
func ClientMetric(db *database.DB, now func() time.Time) func(*web.Ctx) error {
	return func(c *web.Ctx) error {
		raw, err := decodeObject(c)
		if err != nil {
			return err
		}

		if err := rejectUnknown(raw, "name", "value", "page", "locale"); err != nil {
			return err
		}

		row, err := metricReport(raw)
		if err != nil {
			return err
		}

		if err := insertReport(c.Request().Context(), db, row, now().UTC()); err != nil {
			return err
		}

		c.Log().Info("client.web_vital", "name", row.name, "route", row.route)

		return c.NoContent()
	}
}

type clientReport struct {
	kind, platform, appVersion, route, name, message, stack string
	value                                                   *float64
}

func errorReport(raw map[string]json.RawMessage) (clientReport, error) {
	var row clientReport

	version, _, err := jsonString(raw, "appVersion", reportVersionMax)
	if err != nil {
		return row, err
	}

	if err := safeToken(version); err != nil {
		return row, err
	}

	if err := discardScalar(raw, "versionCode"); err != nil {
		return row, err
	}

	if err := discardScalar(raw, "androidApi"); err != nil {
		return row, err
	}

	screen, _, err := jsonString(raw, "screen", reportRouteMax)
	if err != nil {
		return row, err
	}

	if err := routePattern(screen); err != nil {
		return row, err
	}

	message, _, err := jsonString(raw, "message", reportMessageMax)
	if err != nil {
		return row, err
	}

	stack, _, err := jsonString(raw, "stack", reportStackMax)
	if err != nil {
		return row, err
	}

	if message == "" && stack == "" {
		return row, errs.New(http.StatusBadRequest, "Invalid field.")
	}

	row.kind = "error"
	row.platform = "android"
	row.appVersion = version
	row.route = screen
	row.message = message
	row.stack = stack

	return row, nil
}

func metricReport(raw map[string]json.RawMessage) (clientReport, error) {
	var row clientReport

	name, ok, err := jsonString(raw, "name", 16)
	if err != nil {
		return row, err
	}

	if !ok {
		return row, errs.New(http.StatusBadRequest, "Invalid field.")
	}

	if _, known := vitalNames[name]; !known {
		return row, errs.New(http.StatusBadRequest, "Invalid field.")
	}

	value, err := jsonFloat(raw, "value")
	if err != nil {
		return row, err
	}

	page, _, err := jsonString(raw, "page", reportRouteMax)
	if err != nil {
		return row, err
	}

	if err := routePattern(page); err != nil {
		return row, err
	}

	if err := discardScalar(raw, "locale"); err != nil {
		return row, err
	}

	row.kind = "web_vital"
	row.platform = "web"
	row.route = page
	row.name = name
	row.value = &value

	return row, nil
}

func insertReport(ctx context.Context, db *database.DB, row clientReport, at time.Time) error {
	_, err := db.ExecContext(ctx, `INSERT INTO client_reports (
  kind, platform, app_version, route, name, value, message, stack, received_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		row.kind, row.platform, nullIfEmpty(row.appVersion), nullIfEmpty(row.route),
		nullIfEmpty(row.name), row.value, nullIfEmpty(row.message), nullIfEmpty(row.stack), at)
	if err != nil {
		return fmt.Errorf("telemetry: store report: %w", err)
	}

	return nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}

	return s
}

func decodeObject(c *web.Ctx) (map[string]json.RawMessage, error) {
	if !c.IsJSON() {
		return nil, errs.ErrUnsupportedMediaType
	}

	dec := json.NewDecoder(c.Request().Body)

	var raw map[string]json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return nil, errs.ErrRequestTooLarge
		}

		return nil, errs.New(http.StatusBadRequest, "Malformed JSON.")
	}

	if raw == nil {
		return nil, errs.New(http.StatusBadRequest, "Malformed JSON.")
	}

	if dec.More() {
		return nil, errs.New(http.StatusBadRequest, "Request body must contain a single JSON value.")
	}

	return raw, nil
}

func rejectUnknown(raw map[string]json.RawMessage, allow ...string) error {
	ok := make(map[string]struct{}, len(allow))
	for _, key := range allow {
		ok[key] = struct{}{}
	}

	for key := range raw {
		if _, yes := ok[key]; !yes {
			return errs.New(http.StatusBadRequest, "Unknown field.")
		}
	}

	return nil
}

func jsonString(raw map[string]json.RawMessage, key string, maxRunes int) (string, bool, error) {
	v, ok := raw[key]
	if !ok || string(v) == "null" {
		return "", false, nil
	}

	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return "", false, errs.New(http.StatusBadRequest, "Invalid field.")
	}

	if utf8.RuneCountInString(s) > maxRunes {
		return "", false, errs.New(http.StatusBadRequest, "Invalid field.")
	}

	return s, true, nil
}

func jsonFloat(raw map[string]json.RawMessage, key string) (float64, error) {
	v, ok := raw[key]
	if !ok || string(v) == "null" {
		return 0, errs.New(http.StatusBadRequest, "Invalid field.")
	}

	var n float64
	if err := json.Unmarshal(v, &n); err != nil {
		return 0, errs.New(http.StatusBadRequest, "Invalid field.")
	}

	return n, nil
}

// discardScalar accepts a number or a short string and stores nothing.
// versionCode, androidApi and locale are not columns (SDD 12 §5).
func discardScalar(raw map[string]json.RawMessage, key string) error {
	v, ok := raw[key]
	if !ok || string(v) == "null" {
		return nil
	}

	var s string
	if json.Unmarshal(v, &s) == nil {
		if utf8.RuneCountInString(s) > reportVersionMax || strings.Contains(s, "@") || strings.Contains(s, "://") {
			return errs.New(http.StatusBadRequest, "Invalid field.")
		}

		return nil
	}

	var n float64
	if json.Unmarshal(v, &n) == nil {
		return nil
	}

	return errs.New(http.StatusBadRequest, "Invalid field.")
}

func safeToken(s string) error {
	if strings.Contains(s, "@") || strings.Contains(s, "://") {
		return errs.New(http.StatusBadRequest, "Invalid field.")
	}

	return nil
}

// routePattern rejects a full URL or an address. The column keeps a screen
// name or a route pattern only.
func routePattern(s string) error {
	if strings.Contains(s, "://") || strings.Contains(s, "@") || strings.Contains(s, "?") {
		return errs.New(http.StatusBadRequest, "Invalid field.")
	}

	return nil
}
