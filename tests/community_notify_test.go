//go:build integration

package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func TestCommunityNotify(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	pt := makeUser(t, app, "notify-pt@example.com", "user", true)
	out := makeUser(t, app, "notify-out@example.com", "user", true)
	if _, err := db.ExecContext(t.Context(), `UPDATE users SET locale = 'pt-BR' WHERE id = $1;`, pt.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `UPDATE users SET community_mail_opt_out = true WHERE id = $1`,
		out.ID); err != nil {
		t.Fatal(err)
	}

	const applied, rejected = "00000000-0000-7000-8000-0000000000d1", "00000000-0000-7000-8000-0000000000d2"
	for _, s := range []struct {
		id, status string
		author     int64
	}{{applied, "applied", pt.ID}, {rejected, "rejected", out.ID}} {
		if _, err := db.ExecContext(t.Context(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, status, submitted_at)
VALUES ($1, $2, 'edit', 'item', '00000000-0000-0000-0000-0000000000aa', 'movie', 'en', '{}', $3, now())`,
			s.id, s.author, s.status); err != nil {
			t.Fatal(err)
		}
	}

	inbox := func(userID int64, kind string) int {
		var n int
		if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM user_notifications WHERE user_id = $1
  AND kind = $2`, userID, kind).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	// A retry writes one inbox row and one mail, in the author's locale.
	for range 2 {
		if err := jobs.RunCommunityNotify(t.Context(), db, acc,
			community.Notify{SuggestionID: applied, State: "applied"}); err != nil {
			t.Fatal(err)
		}
	}
	if n := inbox(pt.ID, "suggestion_decided"); n != 1 {
		t.Fatalf("inbox rows = %d, want 1", n)
	}
	mails := app.Mails().SentTo(pt.Email)
	if len(mails) != 1 {
		t.Fatalf("mails = %d, want 1", len(mails))
	}
	if !strings.Contains(mails[0].Subject, "sugestão") {
		t.Fatalf("subject not in pt-BR: %q", mails[0].Subject)
	}
	if !strings.Contains(mails[0].HTML, "/suggestions/"+applied) {
		t.Fatalf("mail does not link the suggestion: %s", mails[0].HTML)
	}

	// Opted out: inbox row only.
	if err := jobs.RunCommunityNotify(t.Context(), db, acc,
		community.Notify{SuggestionID: rejected, State: "rejected"}); err != nil {
		t.Fatal(err)
	}
	if n := inbox(out.ID, "suggestion_decided"); n != 1 {
		t.Fatalf("opted-out inbox rows = %d, want 1", n)
	}
	if n := len(app.Mails().SentTo(out.Email)); n != 0 {
		t.Fatalf("opted-out user got %d mails", n)
	}

	// A state without a notice is a no-op.
	if err := jobs.RunCommunityNotify(t.Context(), db, acc,
		community.Notify{SuggestionID: applied, State: "withdrawn"}); err != nil {
		t.Fatal(err)
	}
}

func TestCommunityNotifyOnReject(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedModerationItem(t, app)
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	author := makeUser(t, app, "notify-rej@example.com", "user", true)
	const id = "00000000-0000-7000-8000-0000000000d3"
	seedSuggestion(t, app, id, "submitted", `{"title":{"before":"a","after":"b"}}`)
	if _, err := db.ExecContext(t.Context(), `UPDATE suggestions SET author_id = $2 WHERE id = $1`, id,
		author.ID); err != nil {
		t.Fatal(err)
	}

	admin := browserFor(t, app, makeUser(t, app, "notify-admin@example.com", models.RoleAdmin, true))
	admin.PostForm("/admin/moderation/"+id+"/reject", url.Values{"version": {"1"}, "reason": {"incorrect"}}).
		AssertRedirect("/admin/moderation")

	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM user_notifications WHERE user_id = $1
  AND dedupe_key = $2`, author.ID, "suggestion:"+id+":rejected").Scan(&n); err != nil || n != 1 {
		t.Fatalf("reject inbox rows = %d, %v", n, err)
	}
}
