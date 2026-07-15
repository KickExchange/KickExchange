package engineclient

import "net"

// runWriter is the sole goroutine allowed to write to conn. Everything else
// enqueues onto sendQueue instead of touching the socket directly.
func runWriter(conn net.Conn, sendQueue <-chan []byte, done <-chan struct{}) {
	for {
		select {
		case msg, ok := <-sendQueue:
			if !ok {
				return
			}
			if _, err := conn.Write(msg); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}
