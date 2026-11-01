package catalog

// ItemsChanged is the `catalog.items_changed` job, dispatched with DispatchTx for every committed normalise batch
// (SDD 06 §6 "Events"); search reindexing (SDD 07) and cache purges (SDD 09) hang off it.
type ItemsChanged struct {
	IDs []string `json:"ids"`
}

// Kind names the job on the queue.
func (ItemsChanged) Kind() string { return "catalog.items_changed" }
