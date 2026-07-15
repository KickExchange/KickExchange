package engineclient

import (
	"io"
	"log/slog"
	"net"
)

// AsyncEvent is a push message not tied to any pending request — the
// counterparty side of a trade (request_id == 0), per session.cpp.
type AsyncEvent struct {
	Executed  *ExecutedPayload
	OrderDone *OrderDonePayload
}

func recvExact(conn net.Conn, buf []byte) error {
	_, err := io.ReadFull(conn, buf)
	return err
}

func dispatchAsyncEvent(onAsync func(AsyncEvent), e AsyncEvent, log *slog.Logger, kind string, orderID uint64) {
	if onAsync != nil {
		onAsync(e)
		return
	}
	log.Info("async event (no handler registered)", "kind", kind, "order_id", orderID)
}

// runReader is the sole goroutine reading from conn. It decodes each message
// and either feeds the pending table (request_id != 0) or forwards to the
// async event handler (request_id == 0, e.g. resting-side fills).
func runReader(conn net.Conn, pending *pendingTable, onAsync func(AsyncEvent), log *slog.Logger) error {
	headerBuf := make([]byte, HeaderSize)
	for {
		if err := recvExact(conn, headerBuf); err != nil {
			return err
		}
		h := decodeHeader(headerBuf)

		size, err := wireSize(h.MsgType)
		if err != nil {
			log.Warn("unknown message type from engine", "msg_type", h.MsgType)
			return err
		}
		payload := make([]byte, size)
		if size > 0 {
			if err := recvExact(conn, payload); err != nil {
				return err
			}
		}

		switch h.MsgType {
		case MsgHeartbeat:
			// engine never sends Heartbeat back; nothing to do
		case MsgAccepted:
			if !pending.onAccepted(h.RequestID, decodeAccepted(payload)) {
				log.Warn("Accepted for unknown request_id", "request_id", h.RequestID)
			}
		case MsgRejected:
			if !pending.onRejected(h.RequestID, decodeRejected(payload)) {
				log.Warn("Rejected for unknown request_id", "request_id", h.RequestID)
			}
		case MsgExecuted:
			// Executed never completes a pending Submit (Accepted always
			// precedes it on the wire) - always an async fill notification,
			// either the counterparty's (request_id==0) or a later fill on
			// an order this connection already got Accepted for.
			p := decodeExecuted(payload)
			dispatchAsyncEvent(onAsync, AsyncEvent{Executed: &p}, log, "executed", p.OrderID)
		case MsgOrderDone:
			p := decodeOrderDone(payload)
			if pending.onOrderDone(h.RequestID, p) {
				continue // Cancel's synchronous response
			}
			dispatchAsyncEvent(onAsync, AsyncEvent{OrderDone: &p}, log, "order done", p.OrderID)
		default:
			log.Warn("unexpected message type from engine", "msg_type", h.MsgType)
		}
	}
}
