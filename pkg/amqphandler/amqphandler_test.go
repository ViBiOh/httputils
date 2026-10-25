package amqphandler

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/mock/gomock"
)

func TestHandleMessage(t *testing.T) {
	t.Parallel()

	const tag = uint64(7)

	success := func(context.Context, amqp.Delivery) error { return nil }
	failure := func(context.Context, amqp.Delivery) error { return errors.New("failed") }
	panicking := func(context.Context, amqp.Delivery) error { panic("boom") }

	cases := map[string]struct {
		handler Handler
		headers amqp.Table
		retry   bool
		expect  func(acknowledger *mocks.Acknowledger)
	}{
		"success": {
			success,
			nil,
			true,
			func(acknowledger *mocks.Acknowledger) {
				acknowledger.EXPECT().Ack(tag, false).Return(nil)
			},
		},
		"error with retry": {
			failure,
			nil,
			true,
			func(acknowledger *mocks.Acknowledger) {
				acknowledger.EXPECT().Nack(tag, false, false).Return(nil)
			},
		},
		"error without retry": {
			failure,
			nil,
			false,
			func(acknowledger *mocks.Acknowledger) {
				acknowledger.EXPECT().Ack(tag, false).Return(nil)
			},
		},
		"error with retries exhausted": {
			failure,
			amqp.Table{"x-death": []any{amqp.Table{"count": int64(3)}}},
			true,
			func(acknowledger *mocks.Acknowledger) {
				acknowledger.EXPECT().Ack(tag, false).Return(nil)
			},
		},
		"panic with retry": {
			panicking,
			nil,
			true,
			func(acknowledger *mocks.Acknowledger) {
				acknowledger.EXPECT().Nack(tag, false, false).Return(nil)
			},
		},
		"panic without retry": {
			panicking,
			nil,
			false,
			func(acknowledger *mocks.Acknowledger) {
				acknowledger.EXPECT().Ack(tag, false).Return(nil)
			},
		},
		"panic with retries exhausted": {
			panicking,
			amqp.Table{"x-death": []any{amqp.Table{"count": int64(3)}}},
			true,
			func(acknowledger *mocks.Acknowledger) {
				acknowledger.EXPECT().Ack(tag, false).Return(nil)
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			acknowledger := mocks.NewAcknowledger(ctrl)
			testCase.expect(acknowledger)

			instance := &Service{handler: testCase.handler}
			if testCase.retry {
				instance.retryInterval = time.Second
				instance.maxRetry = 3
			}

			instance.handleMessage(context.Background(), slog.Default(), amqp.Delivery{
				Acknowledger: acknowledger,
				DeliveryTag:  tag,
				Headers:      testCase.headers,
			})
		})
	}
}
