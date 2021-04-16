package comm

import "time"

// Config is the configurations for the comm protocol
type Config struct {
	// Addr is the target address to communicate with
	Addr string
	// DialTimeout sets the timeout when trying to connect to the target address
	DialTimeout time.Duration
	// ReadDeadline sets the deadline for reads using SetReadDeadline
	ReadDeadline time.Time
	// WriteDeadline sets the deadline for writes using SetWriteDeadline
	WriteDeadline time.Time
}

// sets default values
func (c *Config) init() {
	if c.DialTimeout == 0 {
		c.DialTimeout = 30 * time.Second
	}

	if c.ReadDeadline.IsZero() {
		c.ReadDeadline = time.Now().Add(5 * time.Minute)
	}

	if c.WriteDeadline.IsZero() {
		c.WriteDeadline = c.ReadDeadline
	}
}