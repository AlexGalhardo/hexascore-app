package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/signed"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/backup"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestSecurity_A01_BackupOffsiteLink drives GET /backups/offsite/ through the real router (SDD 11 §7.3
// step 6): only a link signed with the APP_KEY-derived offsite key opens, only monthly dumps are served,
// the response is never cached, and a download tool is not turned away by the scrape guard.
func TestSecurity_A01_BackupOffsiteLink(t *testing.T) {
	t.Parallel()

	monthly := backup.ObjectKey("monthly", "test", time.Now().Add(-time.Hour))
	daily := backup.ObjectKey("daily", "test", time.Now().Add(-time.Hour))

	bucket := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/offsite-bucket/"+monthly || r.URL.Path == "/offsite-bucket/"+daily {
			_, _ = w.Write([]byte("sealed frames"))
			return
		}

		http.Error(w, "<Error><Code>NoSuchKey</Code></Error>", http.StatusNotFound)
	}))
	t.Cleanup(bucket.Close)

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"S3_BUCKET": "offsite-bucket", "S3_ENDPOINT": bucket.URL, "S3_USE_PATH_STYLE": "true",
		"S3_INSECURE_HTTP": "true", "S3_ACCESS_KEY_ID": "k", "S3_SECRET_ACCESS_KEY": "s",
	}})

	signer, err := backup.NewOffsiteSigner(app.App)
	if err != nil {
		t.Fatal(err)
	}

	link := func(s *signed.Signer, key string) string {
		t.Helper()

		u, err := storage.TemporaryURL(s, backup.OffsitePrefix, key, time.Hour)
		if err != nil {
			t.Fatal(err)
		}

		return u
	}

	c := app.Client(t).WithHeader("User-Agent", "curl/8.10.1")

	c.Get(link(signer, monthly)).AssertOK().AssertSee("sealed frames").
		AssertHeader("Cache-Control", "private, no-store")

	// Another key (a different APP_KEY or purpose) never opens the dump.
	other := signed.New(bytes.Repeat([]byte{7}, 32))
	c.Get(link(other, monthly)).AssertForbidden()

	c.Get(backup.OffsitePrefix + "/" + monthly).AssertForbidden()
	c.Get(strings.Replace(link(signer, monthly), "signature=", "signature=00", 1)).AssertForbidden()

	// A valid signature outside backups/postgres/monthly/ is still a 404.
	c.Get(link(signer, daily)).AssertNotFound()
	c.Get(link(signer, "covers/x.jpg")).AssertNotFound()
}
