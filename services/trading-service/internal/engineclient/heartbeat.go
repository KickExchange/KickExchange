package engineclient

import "time"

// runHeartbeat enqueues a Heartbeat message every interval. Fire-and-forget —
// the engine's serve_connection loop just `continue`s on Heartbeat, no ack.
func runHeartbeat(interval time.Duration, sendQueue chan<- []byte, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	msg := encodeMessage(Header{Version: ProtocolVersion, MsgType: MsgHeartbeat}, nil)
	for {
		select {
		case <-ticker.C:
			select {
			case sendQueue <- msg:
			case <-done:
				return
			}
		case <-done:
			return
		}
	}
}
