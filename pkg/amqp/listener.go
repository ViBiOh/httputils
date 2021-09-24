package amqp

import (
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/uuid"
)

func (a *Client) getListener() (string, <-chan bool, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

identity:
	name, err := uuid.New()
	if err != nil {
		return "", nil, fmt.Errorf("unable to generate uuid: %s", err)
	}

	if a.listeners[name] != nil {
		goto identity
	}

	listener := make(chan bool)
	a.listeners[name] = listener

	return name, listener, nil
}

func (a *Client) notifyListeners() {
	for _, listener := range a.listeners {
		listener <- true
	}
}

func (a *Client) closeListeners() {
	for name := range a.listeners {
		a.removeListener(name)
	}
}

func (a *Client) removeListener(name string) {
	listener := a.listeners[name]
	if listener == nil {
		return
	}

	close(listener)
	<-listener // drain eventually

	delete(a.listeners, name)
}
