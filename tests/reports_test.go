//go:build integration

package tests

import (
	"net/url"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/notifications"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func (r *listRig) count(q string, args ...any) int {
	r.t.Helper()

	var n int
	if err := r.acc.DB.QueryRowContext(r.t.Context(), q, args...).Scan(&n); err != nil {
		r.t.Fatal(err)
	}

	return n
}

func reportBody(typ, target string) map[string]any {
	return map[string]any{"targetType": typ, "targetId": target, "reason": "spam", "details": "looks like spam"}
}

func TestSecurity_A01_ReportSelf(t *testing.T) {
	r := newListRig(t)
	owner, c := r.user("repself@example.com", "lists:write", "profile:write")
	_, stranger := r.user("repstranger@example.com", "profile:write")
	_, readOnly := r.user("repro@example.com", "catalog:read")

	var list struct {
		ID string `json:"id"`
	}

	c.PostJSON("/app/v1/lists", map[string]any{"type": "game", "title": "mine", "isPublic": false}).JSON(&list)
	if _, err := r.acc.DB.ExecContext(
		t.Context(),
		`UPDATE users SET username = 'selfie' WHERE id = $1`,
		owner.ID,
	); err != nil {
		t.Fatal(err)
	}

	c.PostJSON("/app/v1/reports", reportBody("list", list.ID)).
		AssertForbidden().
		AssertSee("cannot report yourself").
		AssertSee(`"code":"report-self"`)
	c.PostJSON("/app/v1/reports", reportBody("user", "SelFie")).AssertForbidden().AssertSee("cannot report yourself")

	// A private list is invisible to others, so reporting it is a 404, like a missing one.
	stranger.PostJSON("/app/v1/reports", reportBody("list", list.ID)).AssertNotFound()
	stranger.PostJSON("/app/v1/reports", reportBody("user", "nobody")).AssertNotFound()
	readOnly.PostJSON("/app/v1/reports", reportBody("user", "selfie")).AssertForbidden()

	if n := r.count(`SELECT count(*) FROM reports`); n != 0 {
		t.Fatalf("%d reports stored, want 0", n)
	}
}

func TestReportResolveNotifies(t *testing.T) {
	r := newListRig(t)
	owner, c := r.user("repowner@example.com", "lists:write")
	rep1, c1 := r.user("rep1@example.com", "profile:write")
	rep2, c2 := r.user("rep2@example.com", "profile:write")
	admin := makeUser(t, r.app, adminEmail, "admin", true)

	var list struct {
		ID string `json:"id"`
	}

	c.PostJSON("/app/v1/lists", map[string]any{"type": "game", "title": "spammy", "isPublic": true}).JSON(&list)

	c1.PostJSON("/app/v1/reports", reportBody("list", list.ID)).AssertStatus(201)
	// A repeat of an open report is accepted without a second row.
	c1.PostJSON("/app/v1/reports", reportBody("list", list.ID)).AssertStatus(200)
	c2.PostJSON("/app/v1/reports", reportBody("list", list.ID)).AssertStatus(201)

	if n := r.count(`SELECT count(*) FROM reports WHERE reporter_id = $1`, rep1.ID); n != 1 {
		t.Fatalf("reporter 1 has %d rows, want 1", n)
	}

	var id string
	if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT id FROM reports WHERE reporter_id = $1`, rep1.ID).
		Scan(&id); err != nil {
		t.Fatal(err)
	}

	adminC := browserFor(t, r.app, admin)
	browserFor(t, r.app, rep1).PostForm("/admin/reports/"+id+"/resolve", url.Values{"outcome": {"list_hidden"}}).
		AssertNotFound()
	adminC.PostForm("/admin/reports/"+id+"/resolve", url.Values{"outcome": {"user_banned"}}).AssertUnprocessable()
	adminC.PostForm("/admin/reports/"+id+"/resolve", url.Values{"outcome": {"list_hidden"}, "adminNote": {"spam"}}).
		AssertRedirect("/admin/reports/" + id)

	if n := r.count(`SELECT count(*) FROM lists WHERE id = $1 AND hidden`, list.ID); n != 1 {
		t.Fatal("list not hidden")
	}

	// The other open report on the same list is closed alike.
	if n := r.count(`SELECT count(*) FROM reports WHERE status = 'resolved' AND outcome = 'list_hidden'`); n != 2 {
		t.Fatalf("%d reports resolved, want 2", n)
	}

	for _, uid := range []int64{rep1.ID, rep2.ID} {
		if n := r.count(`SELECT count(*) FROM user_notifications WHERE user_id = $1 AND kind = $2`,
			uid, notifications.ReportResolved); n != 1 {
			t.Fatalf("user %d has %d report notices, want 1", uid, n)
		}
	}

	if n := r.count(`SELECT count(*) FROM user_notifications WHERE user_id = $1 AND kind = $2`,
		owner.ID, notifications.ListHidden); n != 1 {
		t.Fatalf("owner has %d list_hidden notices, want 1", n)
	}

	adminC.PostForm("/admin/reports/"+id+"/resolve", url.Values{"outcome": {"no_action"}}).AssertStatus(409)
}

func TestReportResolveBansUser(t *testing.T) {
	r := newListRig(t)
	bad, _ := r.user("badguy@example.com")
	_, c1 := r.user("banrep@example.com", "profile:write")
	admin := makeUser(t, r.app, adminEmail, "admin", true)

	if _, err := r.acc.DB.ExecContext(
		t.Context(),
		`UPDATE users SET username = 'badguy' WHERE id = $1`,
		bad.ID,
	); err != nil {
		t.Fatal(err)
	}

	c1.PostJSON("/app/v1/reports", reportBody("user", "badguy")).AssertStatus(201)

	var id string
	if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT id FROM reports`).Scan(&id); err != nil {
		t.Fatal(err)
	}

	browserFor(t, r.app, admin).PostForm("/admin/reports/"+id+"/resolve", url.Values{"outcome": {"user_banned"}}).
		AssertRedirect("/admin/reports/" + id)

	if !reload(t, r.app, bad.ID).BannedAt.Valid {
		t.Fatal("user not banned")
	}
}

func TestAdminMarksContactHandled(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
	})
	admin := makeUser(t, app, "contact-admin@example.com", "admin", true)
	member := makeUser(t, app, "contact-member@example.com", "user", true)
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	id := uuid.NewV7().String()
	if _, err := db.ExecContext(t.Context(), `INSERT INTO contact_messages (id, name, email, subject, message)
VALUES ($1, 'Ana', 'ana@example.com', 'bug', 'this is a message with more than thirty-two characters')`, id); err != nil {
		t.Fatal(err)
	}

	browserFor(t, app, member).Get("/admin/contact-messages").AssertNotFound()
	browserFor(t, app, admin).Get("/admin/contact-messages").AssertOK()
	browserFor(t, app, admin).PostForm("/admin/contact-messages/"+id, url.Values{"handled": {"false", "true"}})

	var handled bool
	if err := db.QueryRowContext(t.Context(), `SELECT handled FROM contact_messages WHERE id = $1`, id).
		Scan(&handled); err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("handled was not saved")
	}
}

func TestContactHoneypotHTTP(t *testing.T) {
	r := newListRig(t)
	form := url.Values{
		"name": {"Bot"}, "email": {"bot@example.com"}, "subject": {"bug"},
		"message": {"this is a message with more than thirty-two characters"}, "website": {"http://spam"},
	}

	r.app.Client(t).WithHeader("Sec-Fetch-Site", "same-origin").PostForm("/contact", form).AssertStatus(303)

	if n := r.count(`SELECT count(*) FROM contact_messages`); n != 0 {
		t.Fatalf("%d messages stored for a honeypot hit", n)
	}

	// Without the honeypot but with no valid token it is refused, not stored.
	form.Del("website")
	r.app.Client(t).WithHeader("Sec-Fetch-Site", "same-origin").WithHeader("Accept", "application/json").
		PostForm("/contact", form).AssertUnprocessable().AssertSee(`"code":"bot-check-failed"`)
}
