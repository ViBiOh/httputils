package cntxt

import (
	"context"
	"time"
)

var _ context.Context = Context{}

type Context struct {
	context.Context
}

func WithoutDeadline(ctx context.Context) Context {
	return Context{
		ctx,
	}
}

func (cc Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (cc Context) Done() <-chan struct{} {
	return nil
}

func (cc Context) Err() error {
	return nil
}
