package amqp

import (
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/uuid"
)

func (c *Client) getListener() (string, <-chan bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

identity:
	name, err := uuid.New()
	if err != nil {
		return "", nil, fmt.Errorf("unable to generate uuid: %s", err)
	}

	if c.listeners[name] != nil {
		goto identity
	}

	listener := make(chan bool)
	c.listeners[name] = listener

	c.increaseConnection("listener")

	return name, listener, nil
}

func (c *Client) notifyListeners() {
	for _, listener := range c.listeners {
		listener <- true
	}
}

func (c *Client) closeListeners() {
	for name := range c.listeners {
		c.removeListener(name)
	}
}

func (c *Client) removeListener(name string) {
	listener := c.listeners[name]
	if listener == nil {
		return
	}

	close(listener)
	<-listener // drain eventually

	delete(c.listeners, name)
}
