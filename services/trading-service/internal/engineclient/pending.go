package engineclient

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrConnectionLost = errors.New("engineclient: connection lost")
	ErrRequestTimeout = errors.New("engineclient: request timed out waiting for engine response")
)

// SubmitResult carries whichever response completed the request. New*/Modify*
// complete on Accepted; Cancel completes on OrderDone (its only response type
// - session.cpp's handle_cancel_order never sends Accepted). Fills and any
// later OrderDone for a placed order arrive after Submit has already
// returned - see EngineClient.OnAsyncEvent.
type SubmitResult struct {
	Accepted *AcceptedPayload
	Done     *OrderDonePayload
	Rejected *RejectedPayload
}

type pendingRequest struct {
	result SubmitResult
	done   chan error
}

type pendingTable struct {
	mu      sync.Mutex
	nextID  uint64
	entries map[uint64]*pendingRequest
}

func newPendingTable() *pendingTable {
	return &pendingTable{entries: make(map[uint64]*pendingRequest)}
}

func (t *pendingTable) newRequestID() uint64 {
	return atomic.AddUint64(&t.nextID, 1)
}

func (t *pendingTable) register(requestID uint64) *pendingRequest {
	pr := &pendingRequest{done: make(chan error, 1)}
	t.mu.Lock()
	t.entries[requestID] = pr
	t.mu.Unlock()
	return pr
}

func (t *pendingTable) get(requestID uint64) (*pendingRequest, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	pr, ok := t.entries[requestID]
	return pr, ok
}

// complete removes an entry without signalling - used for timeout/error
// cleanup where the caller already knows the outcome via another path.
func (t *pendingTable) complete(requestID uint64) {
	t.mu.Lock()
	delete(t.entries, requestID)
	t.mu.Unlock()
}

// failAll errors out every in-flight request. Called on disconnect so blocked
// Submit/Cancel callers don't hang waiting for a response that'll never come.
func (t *pendingTable) failAll(err error) {
	t.mu.Lock()
	entries := t.entries
	t.entries = make(map[uint64]*pendingRequest)
	t.mu.Unlock()

	for _, pr := range entries {
		pr.done <- err
	}
}

// onAccepted/onRejected/onOrderDone feed a pending request as its completing
// response arrives, from the reader goroutine. Each returns false if
// requestID isn't (or is no longer) pending, telling the reader to route the
// message to the async handler instead.

func (t *pendingTable) onAccepted(requestID uint64, p AcceptedPayload) bool {
	pr, ok := t.get(requestID)
	if !ok {
		return false
	}
	pr.result.Accepted = &p
	t.complete(requestID)
	pr.done <- nil
	return true
}

func (t *pendingTable) onRejected(requestID uint64, p RejectedPayload) bool {
	pr, ok := t.get(requestID)
	if !ok {
		return false
	}
	pr.result.Rejected = &p
	t.complete(requestID)
	pr.done <- nil
	return true
}

func (t *pendingTable) onOrderDone(requestID uint64, p OrderDonePayload) bool {
	pr, ok := t.get(requestID)
	if !ok {
		return false
	}
	pr.result.Done = &p
	t.complete(requestID)
	pr.done <- nil
	return true
}
