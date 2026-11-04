//go:build integration

package llm_test

import (
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/llm"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

const groundedText = "The Matrix is a 1999 science fiction film. It was released in the United States on March 31, 1999, " +
	"and runs 136 minutes. The film follows a hacker who learns the world is a simulation, and it became a landmark " +
	"of action cinema with its bullet-time effects and its philosophical themes."

// A value written in the introduction itself is applied to the item at once; one that is not stays a review.
func TestNormaliseWritesGroundedValuesDirectly(t *testing.T) {
	r := newRig(t, func(int) (int, string) {
		return reply(`{"release_date":"1999-03-31","runtime_minutes":150,"page_count":null}`, 1300, 20)
	})

	film := r.item("movie", groundedText)

	if _, _, err := r.runStep(runAt, 10); err != nil {
		t.Fatal(err)
	}

	if n := r.count(`SELECT count(*) FROM items WHERE id = $1 AND release_date = '1999-03-31'
AND provenance->>'release_date' = 'llm'`, film); n != 1 {
		t.Error("the date is in the text: it must be written to the item with the model as its source")
	}

	if n := r.count(`SELECT count(*) FROM items WHERE id = $1 AND runtime_minutes IS NULL`, film); n != 1 {
		t.Error("150 minutes is not in the text: the item must not get it")
	}

	if n := r.count(
		`SELECT count(*) FROM field_reviews WHERE item_id = $1 AND field = 'runtime_minutes'`,
		film,
	); n != 1 {
		t.Error("the ungrounded runtime must wait for review")
	}
}

func step(r *rig, s run.Step) run.Result {
	r.t.Helper()

	var res run.Result

	if err := r.db.Tx(r.ctx, func(tx *database.Tx) (e error) {
		res, e = s(r.ctx, tx, run.Run{})

		return e
	}); err != nil {
		r.t.Fatal(err)
	}

	return res
}

// The English synopsis is translated into the locales that lack one; an existing synopsis is never replaced.
func TestTranslateFillsMissingSynopses(t *testing.T) {
	pt := strings.Repeat("Um hacker descobre que o mundo é uma simulação. ", 6)
	r := newRig(t, func(int) (int, string) {
		return reply(`{"pt-BR":"`+pt+`","es":null}`, 900, 400)
	})

	film := r.item("movie", groundedText)
	en := strings.Repeat("A hacker learns the world is a simulation. ", 6)

	r.exec(
		`INSERT INTO item_translations (item_id, locale, title, title_source, synopsis, synopsis_source, synopsis_licence,
  synopsis_source_url) VALUES ($1, 'en', 'The Matrix', 'wikipedia', $2, 'wikipedia', 'CC-BY-SA-4.0', 'https://en.wikipedia.org/wiki/The_Matrix'),
($1, 'es', 'Matrix', 'wikipedia', 'Una sinopsis existente.', 'wikipedia', 'CC-BY-SA-4.0', NULL)`,
		film,
		en,
	)

	res := step(r, llm.TranslateStep(r.client, 10, func() time.Time { return runAt }))
	if res.Changed != 1 || res.Failed != 0 {
		t.Fatalf("result = %+v", res)
	}

	if n := r.count(`SELECT count(*) FROM item_translations WHERE item_id = $1 AND locale = 'pt-BR'
AND synopsis_source = 'llm' AND synopsis_licence = 'CC-BY-SA-4.0' AND title = 'The Matrix'`, film); n != 1 {
		t.Error("the pt-BR row must be created with the translation, the original licence and the English title")
	}

	if n := r.count(`SELECT count(*) FROM item_translations WHERE item_id = $1 AND locale = 'es'
AND synopsis = 'Una sinopsis existente.'`, film); n != 1 {
		t.Error("an existing synopsis must not be replaced")
	}

	// asked once: the next run leaves it alone
	if res := step(r, llm.TranslateStep(r.client, 10, func() time.Time { return runAt })); res.Seen != 0 {
		t.Errorf("second run saw %d items, want 0", res.Seen)
	}
}

// Web search is off unless enabled; on, its values always go to review with the cited page.
func TestWebStepIsOffByDefaultAndQueuesReviews(t *testing.T) {
	r := newRig(t, func(int) (int, string) {
		return 200, `{"stop_reason":"end_turn","content":[{"type":"text","text":"{\"release_date\":\"2026-11-19\",\"runtime_minutes\":null,\"page_count\":null}","citations":[{"url":"https://example.org/gta"}]}],"usage":{"input_tokens":2000,"output_tokens":30,"server_tool_use":{"web_search_requests":1}}}`
	})

	r.n++
	game := "01234567-89ab-7def-8123-456789abcdef"
	r.exec(
		`INSERT INTO items (id, type, slug, wikidata_qid, visibility) VALUES ($1, 'game', 'gta-vi', 'Q99', 'public')`,
		game,
	)

	if res := step(r, llm.WebStep(r.client, 10, false, func() time.Time { return runAt })); !res.Done || res.Seen != 0 {
		t.Fatalf("disabled web step = %+v, want done with nothing seen", res)
	}

	res := step(r, llm.WebStep(r.client, 10, true, func() time.Time { return runAt }))
	if res.Changed != 1 {
		t.Fatalf("enabled web step = %+v", res)
	}

	if n := r.count(`SELECT count(*) FROM field_reviews WHERE item_id = $1 AND field = 'release_date'
AND reason LIKE '%https://example.org/gta%'`, game); n != 1 {
		t.Error("the web answer must wait for review with the cited page")
	}

	if n := r.count(`SELECT count(*) FROM items WHERE id = $1 AND release_date IS NULL`, game); n != 1 {
		t.Error("a web answer is never written to the item directly")
	}

	if n := r.count(`SELECT count(*) FROM llm_usage WHERE job = 'llm.websearch' AND cost_usd >= 0.01`); n != 1 {
		t.Error("the search must be billed in the usage row")
	}
}
