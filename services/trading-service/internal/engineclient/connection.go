package engineclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

var ErrVersionMismatch = errors.New("engineclient: protocol version rejected by engine")

const (
	backoffInitial = 1 * time.Second
	backoffMax     = 30 * time.Second
)

// connState holds everything tied to one live TCP connection. Swapped out
// wholesale on reconnect.
type connState struct {
	conn      net.Conn
	sendQueue chan []byte
	done      chan struct{}
	wg        sync.WaitGroup
}

func (c *Client) dialAndHandshake(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", c.cfg.EngineHost, c.cfg.EnginePort)
	c.setStatus(StatusConnecting)
	c.log.Info("connecting", "addr", addr)

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		c.setStatus(StatusDisconnected)
		return err
	}
	c.log.Info("connected", "addr", addr)
	c.setStatus(StatusHandshaking)

	hello := encodeMessage(Header{Version: ProtocolVersion, MsgType: MsgHello},
		encodeHello(HelloPayload{MinVersion: c.cfg.ProtocolVersion, MaxVersion: c.cfg.ProtocolVersion}))
	if _, err := conn.Write(hello); err != nil {
		conn.Close()
		c.setStatus(StatusDisconnected)
		return err
	}

	headerBuf := make([]byte, HeaderSize)
	if err := recvExact(conn, headerBuf); err != nil {
		conn.Close()
		c.setStatus(StatusDisconnected)
		return err
	}
	h := decodeHeader(headerBuf)
	if h.MsgType != MsgHelloAck {
		conn.Close()
		c.setStatus(StatusDisconnected)
		return fmt.Errorf("engineclient: expected HelloAck, got msg_type %d", h.MsgType)
	}
	ackBuf := make([]byte, 9)
	if err := recvExact(conn, ackBuf); err != nil {
		conn.Close()
		c.setStatus(StatusDisconnected)
		return err
	}
	ack := decodeHelloAck(ackBuf)
	if ack.Ok == 0 {
		conn.Close()
		c.setStatus(StatusDisconnected)
		return ErrVersionMismatch
	}
	c.log.Info("HELLO OK", "accepted_version", ack.AcceptedVersion, "engine_version", ack.EngineVersion)
	c.setStatus(StatusReady)

	st := &connState{
		conn:      conn,
		sendQueue: make(chan []byte, 16),
		done:      make(chan struct{}),
	}
	st.wg.Add(2)
	go func() {
		defer st.wg.Done()
		runWriter(st.conn, st.sendQueue, st.done)
	}()
	go func() {
		defer st.wg.Done()
		err := runReader(st.conn, c.pending, c.dispatchAsync, c.log)
		if err != nil {
			c.log.Warn("disconnected", "error", err)
		}
		close(st.done)
	}()
	go func() {
		runHeartbeat(c.cfg.HeartbeatInterval, st.sendQueue, st.done)
	}()

	c.mu.Lock()
	c.state = st
	c.mu.Unlock()

	c.log.Info("Heartbeat Started", "interval", c.cfg.HeartbeatInterval)
	return nil
}

// connectWithRetry retries dialAndHandshake with exponential backoff on
// transport errors. A version mismatch is not transient - it bubbles up
// immediately without retrying.
func (c *Client) connectWithRetry(ctx context.Context) error {
	backoff := backoffInitial
	for {
		err := c.dialAndHandshake(ctx)
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrVersionMismatch) {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		c.log.Warn("connect failed, retrying", "error", err, "backoff", backoff)
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
		backoff *= 2
		if backoff > backoffMax {
			backoff = backoffMax
		}
	}
}

// superviseReconnects waits for the current connection to drop, fails any
// in-flight requests, then reconnects with backoff. Runs until ctx is
// cancelled or a version mismatch makes reconnecting pointless.
func (c *Client) superviseReconnects(ctx context.Context) {
	for {
		c.mu.Lock()
		st := c.state
		c.mu.Unlock()
		if st == nil {
			return
		}

		<-st.done
		st.wg.Wait()
		c.setStatus(StatusDisconnected)
		if ctx.Err() != nil {
			return
		}

		c.pending.failAll(ErrConnectionLost)
		c.log.Warn("reconnecting")
		if err := c.connectWithRetry(ctx); err != nil {
			c.log.Error("reconnect failed permanently", "error", err)
			return
		}
		c.log.Info("reconnected")
	}
}
