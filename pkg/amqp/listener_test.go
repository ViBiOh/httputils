package amqp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestClient() *Client {
	return &Client{listeners: make(map[string]*listener)}
}

func newStartedListener(client *Client) *listener {
	item := client.createListener(context.Background(), 1)
	close(item.done)

	return item
}

func TestClose(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		listeners int
		closes    int
	}{
		"no listener": {
			0,
			1,
		},
		"listeners without channel": {
			3,
			1,
		},
		"closed twice": {
			2,
			2,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			client := newTestClient()

			for range testCase.listeners {
				newStartedListener(client)
			}

			for range testCase.closes {
				client.Close(context.Background())
			}

			assert.Empty(t, client.listeners)
			assert.True(t, client.closed)
		})
	}
}

func TestReconnectListeners(t *testing.T) {
	t.Parallel()

	t.Run("after close", func(t *testing.T) {
		t.Parallel()

		client := newTestClient()
		newStartedListener(client)

		client.Close(context.Background())

		assert.NotPanics(t, func() {
			client.reconnectListeners(context.Background())
		})
	})

	t.Run("without connection", func(t *testing.T) {
		t.Parallel()

		client := newTestClient()
		item := newStartedListener(client)

		assert.NotPanics(t, func() {
			client.reconnectListeners(context.Background())
		})
		assert.Empty(t, item.reconnect)
	})

	t.Run("signals once even when called again", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		connection := mocks.NewAMQPConnection(ctrl)
		connection.EXPECT().Channel().Return(nil, errors.New("down")).Times(2)

		client := newTestClient()
		client.connection = connection

		item := newStartedListener(client)

		done := make(chan struct{})
		go func() {
			defer close(done)

			client.reconnectListeners(context.Background())
			client.reconnectListeners(context.Background())
		}()

		select {
		case <-done:
		case <-time.After(time.Second * 5):
			t.Fatal("reconnectListeners blocked on a pending reconnection")
		}

		assert.Len(t, item.reconnect, 1)
	})
}

func TestReconnectWhenClosed(t *testing.T) {
	t.Parallel()

	client := newTestClient()
	client.closed = true

	assert.NoError(t, client.reconnect(context.Background()))
	assert.Nil(t, client.connection)
}

func TestListenFailure(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	connection := mocks.NewAMQPConnection(ctrl)
	connection.EXPECT().Channel().Return(nil, errors.New("down"))

	client := newTestClient()
	client.connection = connection

	_, _, err := client.Listen(func() (string, error) { return "queue", nil }, "exchange", "key")

	require.Error(t, err)
	assert.Empty(t, client.listeners)

	assert.NoError(t, client.StopListener("unknown"))
}

func TestListenerShutdown(t *testing.T) {
	t.Parallel()

	client := newTestClient()
	item := newStartedListener(client)

	assert.NoError(t, item.shutdown())
	assert.NoError(t, item.shutdown())
}

func TestStopListener(t *testing.T) {
	t.Parallel()

	client := newTestClient()
	item := newStartedListener(client)

	assert.NoError(t, client.StopListener(item.name))
	assert.Empty(t, client.listeners)

	assert.NoError(t, client.StopListener(item.name))
}

func TestForwardStopsWhenConsumerIsGone(t *testing.T) {
	t.Parallel()

	client := newTestClient()
	item := client.createListener(context.Background(), 1)

	input := make(chan amqp.Delivery, 1)
	output := make(chan amqp.Delivery)

	go client.forward(context.Background(), item, func() (string, error) { return "queue", nil }, input, output, "exchange", "key")

	input <- amqp.Delivery{}

	require.NoError(t, item.cancel())

	close(input)

	select {
	case <-item.done:
	case <-time.After(time.Second * 5):
		t.Fatal("forward is still blocked on a consumer that stopped reading")
	}
}
