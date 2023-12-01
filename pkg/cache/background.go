package cache

import (
	"context"
	"log/slog"
	"time"
)

var asyncActionTimeout = time.Second * 5

func doInBackground(ctx context.Context, callback func(ctx context.Context) error) {
	ctx, cancel := context.WithTimeout(ctx, asyncActionTimeout)
	defer cancel()

	if err := callback(ctx); err != nil {
		slog.ErrorContext(ctx, "background callback", "err", err)
	}
}
