// Package pricefeed is an in-process pub/sub for last-trade prices per
// asset, fed by the engine's async Executed events. No Redis/NATS needed
// at single-instance scale.
package pricefeed

import "sync"

type Feed struct {
	mu   sync.Mutex
	subs map[uint64]map[chan float64]struct{}
}

func New() *Feed {
	return &Feed{subs: make(map[uint64]map[chan float64]struct{})}
}

// Publish is called from the engineclient reader goroutine - sends are
// non-blocking so a slow GraphQL subscriber can never stall the engine
// connection.
func (f *Feed) Publish(assetID uint64, price float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for ch := range f.subs[assetID] {
		select {
		case ch <- price:
		default:
		}
	}
}

// Subscribe returns a channel of price updates for assetID and an
// unsubscribe func the caller must invoke exactly once when done.
func (f *Feed) Subscribe(assetID uint64) (<-chan float64, func()) {
	ch := make(chan float64, 1)

	f.mu.Lock()
	if f.subs[assetID] == nil {
		f.subs[assetID] = make(map[chan float64]struct{})
	}
	f.subs[assetID][ch] = struct{}{}
	f.mu.Unlock()

	unsubscribe := func() {
		f.mu.Lock()
		delete(f.subs[assetID], ch)
		if len(f.subs[assetID]) == 0 {
			delete(f.subs, assetID)
		}
		f.mu.Unlock()
		close(ch)
	}
	return ch, unsubscribe
}
