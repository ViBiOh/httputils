package amqp

import "errors"

// Enabled checks if connection is setup
func (a *Client) Enabled() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.connection != nil
}

// Ping checks if connection is live
func (a *Client) Ping() error {
	if !a.Enabled() {
		return nil
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if a.connection.IsClosed() {
		return errors.New("amqp client closed")
	}

	return nil
}

// Vhost returns connection Vhost
func (a *Client) Vhost() string {
	return a.vhost
}
