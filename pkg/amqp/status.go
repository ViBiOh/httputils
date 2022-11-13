package amqp

import "errors"

func (c *Client) Enabled() bool {
	c.mutex.RLock()
	enabled := c.connection != nil
	c.mutex.RUnlock()

	return enabled
}

func (c *Client) Ping() error {
	c.mutex.RLock()

	if c.connection != nil && c.connection.IsClosed() {
		return errors.New("amqp client closed")
	}

	c.mutex.RUnlock()

	return nil
}

func (c *Client) Vhost() string {
	return c.vhost
}
