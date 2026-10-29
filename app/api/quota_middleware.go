package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// statusWriter records the status the handler wrote.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	return w.ResponseWriter.Write(b)
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// Quota meters /api/v1 requests after Guard (SDD 05 §7.4): it sets X-Quota-Limit, X-Quota-Remaining and
// X-Quota-Reset, answers 402 quota-exhausted when nothing is left, and refunds a 400 or 5xx answer (FR-05-4).
// The plan's own resolver error or a stale meter answers 503: fail closed.
func Quota(m *Meter, plans PlanResolver, problemBase string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := web.NewCtx(w, r, web.ServicesFrom(r.Context()))

			u, err := auth.User[models.User](c, GuardName)
			if err != nil {
				c.HandleError(err)
				return
			}

			keyID, _ := KeyID(c)

			plan, err := plans(r.Context(), u.ID)
			if err == nil {
				var usage Usage

				usage, err = m.Admit(r.Context(), u.ID, keyID, plan)
				if err == nil {
					serve(w, r, next, m, usage, problemBase)
					return
				}
			}

			if errors.Is(err, ErrMeterStale) {
				web.SecurityEvent(r, "api.quota_stale")
			}

			c.HandleError(Unavailable(err))
		})
	}
}

func serve(w http.ResponseWriter, r *http.Request, next http.Handler, m *Meter, usage Usage, problemBase string) {
	h := w.Header()
	h.Set("X-Quota-Limit", strconv.Itoa(usage.Plan.Quota))
	h.Set("X-Quota-Remaining", strconv.Itoa(usage.Remaining))
	h.Set("X-Quota-Reset", strconv.Itoa(usage.ResetIn))

	if usage.Exhausted {
		body, _ := json.Marshal(map[string]any{
			"type":   strings.TrimSuffix(problemBase, "/") + "/quota-exhausted",
			"title":  "Quota exhausted",
			"status": http.StatusPaymentRequired,
			"code":   "quota-exhausted",
			"plan":   usage.Plan.ID,
			"limit":  usage.Plan.Quota,
		})
		h.Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write(body)

		return
	}

	sw := &statusWriter{ResponseWriter: w}

	defer func() {
		// A panic is answered 500 by the recover middleware: refund it like any 5xx.
		if rec := recover(); rec != nil {
			m.Refund(usage, true)
			panic(rec)
		}

		if sw.status == http.StatusBadRequest || sw.status >= http.StatusInternalServerError {
			m.Refund(usage, sw.status >= http.StatusInternalServerError)
		}
	}()

	next.ServeHTTP(sw, r)
}
