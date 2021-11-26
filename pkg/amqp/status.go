package amqp

import "errors"

// Enabled checks if connection is setup
func (c *Client) Enabled() bool {
	c.RLock()
	defer c.RUnlock()

	return c.connection != nil
}

// Ping checks if connection is live
func (c *Client) Ping() error {
	if !c.Enabled() {
		return nil
	}

	c.RLock()
	defer c.RUnlock()

	if c.connection.IsClosed() {
		return errors.New("amqp client closed")
	}

	return nil
}

// Vhost returns connection Vhost
func (c *Client) Vhost() string {
	return c.vhost
}
