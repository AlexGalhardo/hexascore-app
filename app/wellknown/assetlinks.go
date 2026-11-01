// Package wellknown serves the /.well-known documents other platforms fetch from the site.
package wellknown

import (
	"encoding/json"
	"strings"

	"github.com/AlexGalhardo/vukem/web"
)

type target struct {
	Namespace    string   `json:"namespace"`
	PackageName  string   `json:"package_name"`
	Fingerprints []string `json:"sha256_cert_fingerprints"`
}

type statement struct {
	Relation []string `json:"relation"`
	Target   target   `json:"target"`
}

// AssetLinks serves GET /.well-known/assetlinks.json for verified Android App Links
// (SDD 10 §7.6). certs is ANDROID_CERT_SHA256: a comma-separated list (upload key,
// later the Play signing key). With no fingerprint the body is [] so nothing verifies.
func AssetLinks(pkg, certs string) web.HandlerFunc {
	statements := []statement{}

	var prints []string
	for f := range strings.SplitSeq(certs, ",") {
		if f = strings.TrimSpace(f); f != "" {
			prints = append(prints, f)
		}
	}

	if pkg != "" && len(prints) > 0 {
		statements = append(statements, statement{
			Relation: []string{"delegate_permission/common.handle_all_urls"},
			Target:   target{Namespace: "android_app", PackageName: pkg, Fingerprints: prints},
		})
	}

	body, err := json.Marshal(statements)

	return func(c *web.Ctx) error {
		if err != nil {
			return err
		}

		c.Writer().Header().Set("Content-Type", "application/json")
		c.Writer().Header().Set("Cache-Control", "public, max-age=3600")
		_, werr := c.Writer().Write(body)

		return werr
	}
}
