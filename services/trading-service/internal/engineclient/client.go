package engineclient

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"kickexchange/trading-service/internal/config"
)

// Bounds Submit/Cancel wait time; failAll only covers disconnects, not a stuck live request.
const defaultRequestTimeout = 5 * time.Second

// EngineClient is the only interface the rest of the service should depend
// on. Nobody outside this package sees a socket.
type EngineClient interface {
	Connect(ctx context.Context) error
	Close()

	SubmitLimit(assetID uint64, p NewLimitPayload) (SubmitResult, error)
	SubmitMarket(assetID uint64, p NewMarketPayload) (SubmitResult, error)
	SubmitStop(assetID uint64, p NewStopPayload) (SubmitResult, error)
	SubmitStopLimit(assetID uint64, p NewStopLimitPayload) (SubmitResult, error)
	ModifyLimit(assetID uint64, p ModifyLimitPayload) (SubmitResult, error)
	ModifyStop(assetID uint64, p ModifyStopPayload) (SubmitResult, error)
	ModifyStopLimit(assetID uint64, p ModifyStopLimitPayload) (SubmitResult, error)
	Cancel(assetID uint64, orderID uint64) (SubmitResult, error)

	// OnAsyncEvent registers the handler for request_id==0 pushes (fills on
	// the resting side of a trade this connection didn't initiate). Not
	// called by anything this PR - main.go only exercises Connect/Heartbeat.
	OnAsyncEvent(func(AsyncEvent))
}

type Client struct {
	cfg config.Config
	log *slog.Logger

	pending *pendingTable

	mu     sync.Mutex
	state  *connState
	status atomic.Int32

	onAsyncMu sync.RWMutex
	onAsync   func(AsyncEvent)
}

func New(cfg config.Config, log *slog.Logger) *Client {
	return &Client{
		cfg:     cfg,
		log:     log,
		pending: newPendingTable(),
	}
}

func (c *Client) OnAsyncEvent(f func(AsyncEvent)) {
	c.onAsyncMu.Lock()
	c.onAsync = f
	c.onAsyncMu.Unlock()
}

// dispatchAsync is passed to runReader as a stable func value so OnAsyncEvent
// can be registered/changed without restarting the reader.
func (c *Client) dispatchAsync(e AsyncEvent) {
	c.onAsyncMu.RLock()
	f := c.onAsync
	c.onAsyncMu.RUnlock()
	if f != nil {
		f(e)
	}
}

// Connect performs the initial handshake, blocking on transport-level retries
// (e.g. the engine container isn't accepting connections yet). A protocol
// version mismatch is returned immediately since retrying won't fix it -
// callers should treat that as fatal. Once connected, disconnects are handled
// by a background reconnect loop and do not surface here.
func (c *Client) Connect(ctx context.Context) error {
	if err := c.connectWithRetry(ctx); err != nil {
		return err
	}
	go c.superviseReconnects(ctx)
	return nil
}

func (c *Client) Close() {
	c.setStatus(StatusClosing)
	c.mu.Lock()
	st := c.state
	c.mu.Unlock()
	if st == nil {
		c.setStatus(StatusDisconnected)
		return
	}
	st.conn.Close()
	<-st.done
	st.wg.Wait()
	c.setStatus(StatusDisconnected)
}

func (c *Client) send(msgType MsgType, requestID uint64, assetID uint64, payload []byte) error {
	c.mu.Lock()
	st := c.state
	c.mu.Unlock()
	if st == nil {
		return ErrConnectionLost
	}
	msg := encodeMessage(Header{
		Version:   ProtocolVersion,
		MsgType:   msgType,
		RequestID: requestID,
		AssetID:   assetID,
	}, payload)

	select {
	case st.sendQueue <- msg:
		return nil
	case <-st.done:
		return ErrConnectionLost
	}
}

func (c *Client) submit(msgType MsgType, assetID uint64, payload []byte) (SubmitResult, error) {
	requestID := c.pending.newRequestID()
	pr := c.pending.register(requestID)

	if err := c.send(msgType, requestID, assetID, payload); err != nil {
		c.pending.complete(requestID)
		return SubmitResult{}, err
	}

	select {
	case err := <-pr.done:
		if err != nil {
			return SubmitResult{}, err
		}
	case <-time.After(defaultRequestTimeout):
		c.pending.complete(requestID)
		return SubmitResult{}, ErrRequestTimeout
	}
	if pr.result.Rejected != nil {
		return pr.result, fmt.Errorf("engineclient: rejected, reason=%d", pr.result.Rejected.Reason)
	}
	return pr.result, nil
}

func (c *Client) SubmitLimit(assetID uint64, p NewLimitPayload) (SubmitResult, error) {
	return c.submit(MsgNewLimit, assetID, encodeNewLimit(p))
}

func (c *Client) SubmitMarket(assetID uint64, p NewMarketPayload) (SubmitResult, error) {
	return c.submit(MsgNewMarket, assetID, encodeNewMarket(p))
}

func (c *Client) SubmitStop(assetID uint64, p NewStopPayload) (SubmitResult, error) {
	return c.submit(MsgNewStop, assetID, encodeNewStop(p))
}

func (c *Client) SubmitStopLimit(assetID uint64, p NewStopLimitPayload) (SubmitResult, error) {
	return c.submit(MsgNewStopLimit, assetID, encodeNewStopLimit(p))
}

func (c *Client) ModifyLimit(assetID uint64, p ModifyLimitPayload) (SubmitResult, error) {
	return c.submit(MsgModifyLimit, assetID, encodeModifyLimit(p))
}

func (c *Client) ModifyStop(assetID uint64, p ModifyStopPayload) (SubmitResult, error) {
	return c.submit(MsgModifyStop, assetID, encodeModifyStop(p))
}

func (c *Client) ModifyStopLimit(assetID uint64, p ModifyStopLimitPayload) (SubmitResult, error) {
	return c.submit(MsgModifyStopLimit, assetID, encodeModifyStopLimit(p))
}

func (c *Client) Cancel(assetID uint64, orderID uint64) (SubmitResult, error) {
	return c.submit(MsgCancelOrder, assetID, encodeCancelOrder(CancelOrderPayload{OrderID: orderID}))
}

var _ EngineClient = (*Client)(nil)
