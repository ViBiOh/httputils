package amqp

import "errors"

func (c *Client) Enabled() bool {
	c.RLock()
	enabled := c.connection != nil
	c.RUnlock()

	return enabled
}

func (c *Client) Ping() error {
	c.RLock()

	if c.connection != nil && c.connection.IsClosed() {
		return errors.New("amqp client closed")
	}

	c.RUnlock()

	return nil
}

func (c *Client) Vhost() string {
	return c.vhost
}
