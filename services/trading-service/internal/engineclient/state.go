package engineclient

// Status is the connection lifecycle state, exposed for logging/debugging.
type Status int32

const (
	StatusDisconnected Status = iota
	StatusConnecting
	StatusHandshaking
	StatusReady
	StatusClosing
)

func (s Status) String() string {
	switch s {
	case StatusDisconnected:
		return "disconnected"
	case StatusConnecting:
		return "connecting"
	case StatusHandshaking:
		return "handshaking"
	case StatusReady:
		return "ready"
	case StatusClosing:
		return "closing"
	default:
		return "unknown"
	}
}

func (c *Client) setStatus(s Status) {
	c.status.Store(int32(s))
}

func (c *Client) Status() Status {
	return Status(c.status.Load())
}
