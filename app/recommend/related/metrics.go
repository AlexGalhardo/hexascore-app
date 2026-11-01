package related

import (
	"sync"
)

// Metrics counts reco.impressions and reco.clicks per surface and item type (SDD 08 section 8).
// shortcut: in-process counters with no exporter, since the app has no metrics backend yet; wire them to it then.
type Metrics struct {
	mu       sync.Mutex
	imp, clk map[string]int64
}

// NewMetrics returns empty counters.
func NewMetrics() *Metrics { return &Metrics{imp: map[string]int64{}, clk: map[string]int64{}} }

func (m *Metrics) add(c map[string]int64, surface, typ string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c[surface+"|"+typ]++
}

func (m *Metrics) get(c map[string]int64, surface, typ string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	return c[surface+"|"+typ]
}

// Impression records one served block.
func (m *Metrics) Impression(surface, typ string) { m.add(m.imp, surface, typ) }

// Click records one visit that arrived through the block; surfaces other than similar and crossmedia are ignored.
func (m *Metrics) Click(surface, typ string) {
	if FromSurface(surface) != "" {
		m.add(m.clk, surface, typ)
	}
}

// Impressions returns the impression count.
func (m *Metrics) Impressions(surface, typ string) int64 { return m.get(m.imp, surface, typ) }

// Clicks returns the click count.
func (m *Metrics) Clicks(surface, typ string) int64 { return m.get(m.clk, surface, typ) }

// FromSurface returns the ?from= value when it names an item-page block, else "".
func FromSurface(from string) string {
	if from == "similar" || from == "crossmedia" {
		return from
	}

	return ""
}
