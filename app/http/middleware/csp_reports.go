package middleware

import (
	"io"

	"github.com/AlexGalhardo/vukem/web"
)

// CSPReport accepts a browser CSP violation report. The body is already capped
// at 16 KiB. The report is logged and discarded; it is not stored.
func CSPReport(c *web.Ctx) error {
	n, err := io.Copy(io.Discard, c.Request().Body)
	if err != nil {
		return err
	}

	c.Log().Info("csp report", "bytes", n)

	return c.NoContent()
}
