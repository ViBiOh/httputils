package tracer

import (
	"context"
	"time"
)

var _ context.Context = contextClone{}

type contextClone struct {
	context.Context
}

func CloneContext(ctx context.Context) contextClone {
	return contextClone{
		ctx,
	}
}

func (cc contextClone) Deadline() (deadline time.Time, ok bool) {
	return
}

func (cc contextClone) Done() <-chan struct{} {
	return nil
}

func (cc contextClone) Err() error {
	return nil
}
