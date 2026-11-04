package run

import "testing"

func TestQueueForRoutesOneCategoryToItsLane(t *testing.T) {
	t.Cleanup(func() { SetLanes(false) })

	movie := []byte(`{"types":["movie"],"chain":{"name":"x","stage":0}}`)
	two := []byte(`{"types":["movie","series"]}`)
	odd := []byte(`{"types":["podcast"]}`)

	SetLanes(false)

	if got := QueueFor(movie); got != Queue {
		t.Errorf("lanes off: %s, want %s", got, Queue)
	}

	SetLanes(true)

	for params, want := range map[string]string{
		string(movie): "ingest-movie", string(two): Queue, string(odd): Queue, "": Queue, "{": Queue,
	} {
		if got := QueueFor([]byte(params)); got != want {
			t.Errorf("lanes on, %q: %s, want %s", params, got, want)
		}
	}
}
