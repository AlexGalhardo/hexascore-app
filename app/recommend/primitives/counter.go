// Package primitives holds the anonymous activity counters of SDD 08 section 7.1: views and search clicks counted
// in memory, deduplicated per visitor and UTC day, and flushed to item_activity_daily once a minute.
package primitives

import (
	"container/list"
	"sync"
	"sync/atomic"
)

// MaxKeys caps the buffered counter (SDD 08 section 8): above it the oldest keys are dropped.
const MaxKeys = 100_000

// Key identifies one counter row. Ref is an item slug (web pages), an item id or slug (/app/v1 reads) or an item id
// (search clicks, Type empty); the flush resolves it against items. Day is the UTC day, "2006-01-02".
type Key struct{ Type, Ref, Day string }

// Tally is what was counted for a Key since the last flush.
type Tally struct{ Views, Clicks int }

type slot struct {
	key   Key
	tally Tally
}

// Counter is a bounded, insertion-ordered tally map.
type Counter struct {
	mu      sync.Mutex
	max     int
	order   *list.List // front is the oldest
	items   map[Key]*list.Element
	dropped atomic.Int64
	onDrop  func(n int)
}

// NewCounter returns a Counter holding at most max keys; onDrop (optional) is told how many keys a cap hit dropped.
func NewCounter(limit int, onDrop func(n int)) *Counter {
	return &Counter{max: limit, order: list.New(), items: map[Key]*list.Element{}, onDrop: onDrop}
}

// Add adds views and clicks to k.
func (c *Counter) Add(k Key, views, clicks int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[k]; ok {
		s := slotOf(el)
		s.tally.Views += views
		s.tally.Clicks += clicks

		return
	}

	c.items[k] = c.order.PushBack(&slot{key: k, tally: Tally{views, clicks}})
	c.trim()
}

// Swap takes every count and starts empty.
func (c *Counter) Swap() map[Key]Tally {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make(map[Key]Tally, len(c.items))
	for k, el := range c.items {
		out[k] = slotOf(el).tally
	}

	c.order.Init()
	c.items = map[Key]*list.Element{}

	return out
}

// Merge adds counts back after a failed flush. Keys not seen since the swap go to the old end, so a cap hit drops them
// before anything newer.
func (c *Counter) Merge(batch map[Key]Tally) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, t := range batch {
		if el, ok := c.items[k]; ok {
			s := slotOf(el)
			s.tally.Views += t.Views
			s.tally.Clicks += t.Clicks

			continue
		}

		c.items[k] = c.order.PushFront(&slot{key: k, tally: t})
	}

	c.trim()
}

// Len is the number of buffered keys.
func (c *Counter) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.items)
}

// Dropped is the total number of keys dropped by the cap.
func (c *Counter) Dropped() int64 { return c.dropped.Load() }

// trim drops the oldest keys above the cap. The caller holds mu.
func (c *Counter) trim() {
	n := 0
	for len(c.items) > c.max {
		front := c.order.Front()
		delete(c.items, slotOf(front).key)
		c.order.Remove(front)
		n++
	}

	if n > 0 {
		c.dropped.Add(int64(n))

		if c.onDrop != nil {
			c.onDrop(n)
		}
	}
}

func slotOf(el *list.Element) *slot {
	s, ok := el.Value.(*slot)
	if !ok {
		return &slot{}
	}

	return s
}
