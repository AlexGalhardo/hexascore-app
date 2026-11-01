package tests

import (
	"encoding/json"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"
	"github.com/google/go-cmp/cmp"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestAssetLinks(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"ANDROID_PACKAGE":     "app.hexascore",
		"ANDROID_CERT_SHA256": " AA:BB:CC , DD:EE:FF ,",
	}})
	res := app.Client(t).Get("/.well-known/assetlinks.json").AssertOK().
		AssertHeader("Content-Type", "application/json").
		AssertHeader("Cache-Control", "public, max-age=3600")

	var body []map[string]any
	if err := json.Unmarshal([]byte(res.Body()), &body); err != nil {
		t.Fatal(err)
	}

	want := []map[string]any{{
		"relation": []any{"delegate_permission/common.handle_all_urls"},
		"target": map[string]any{
			"namespace":                "android_app",
			"package_name":             "app.hexascore",
			"sha256_cert_fingerprints": []any{"AA:BB:CC", "DD:EE:FF"},
		},
	}}
	if diff := cmp.Diff(want, body); diff != "" {
		t.Errorf("assetlinks (-want +got):\n%s", diff)
	}
}

// Without fingerprints there is nothing to verify: an empty list, never a 500 or a redirect.
func TestAssetLinks_Unconfigured(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{})
	res := app.Client(t).Get("/.well-known/assetlinks.json").AssertOK()

	if got := res.Body(); got != "[]" {
		t.Errorf("body = %s", got)
	}
}
