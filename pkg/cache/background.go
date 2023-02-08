package cache

import (
	"context"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

var asyncActionTimeout = time.Second * 5

func doInBackground(ctx context.Context, name string, callback func(ctx context.Context) error) {
	ctx, cancel := context.WithTimeout(ctx, asyncActionTimeout)
	defer cancel()

	if err := callback(ctx); err != nil {
		logger.Error("%s: %s", name, err)
	}
}
