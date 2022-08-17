package amqp

import "errors"

// Enabled checks if connection is set up.
func (c *Client) Enabled() bool {
	c.RLock()
	enabled := c.connection != nil
	c.RUnlock()

	return enabled
}

// Ping checks if connection is live.
func (c *Client) Ping() error {
	c.RLock()

	if c.connection != nil && c.connection.IsClosed() {
		return errors.New("amqp client closed")
	}

	c.RUnlock()

	return nil
}

// Vhost returns connection Vhost.
func (c *Client) Vhost() string {
	return c.vhost
}
