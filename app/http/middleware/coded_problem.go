package middleware

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"
)

// CodedProblem wraps a handler so a JSON client gets an RFC 9457 problem
// whose type and `code` carry the kebab-case code that services put in
// errs.HTTPError Params["code"]. workaround(vukem): VF-016/VF-039: errs.HTTPError
// has no extension members, so the default handler answers about:blank.
func CodedProblem(problemBase string, next func(*web.Ctx) error) func(*web.Ctx) error {
	return func(c *web.Ctx) error {
		err := next(c)
		if err == nil || !c.WantsJSON() {
			return err
		}

		he, ok := errors.AsType[*errs.HTTPError](err)
		if !ok {
			return err
		}

		code, _ := he.Params["code"].(string)
		if code == "" {
			return err
		}

		raw, mErr := json.Marshal(map[string]any{
			"type":   strings.TrimSuffix(problemBase, "/") + "/" + code,
			"title":  he.LocalizedText(c.Context()),
			"status": he.Status,
			"code":   code,
		})
		if mErr != nil {
			return mErr
		}

		for k, vs := range he.Headers {
			for _, v := range vs {
				c.SetHeader(k, v)
			}
		}

		return c.Blob(he.Status, "application/problem+json", raw)
	}
}
