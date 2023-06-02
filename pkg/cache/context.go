package cache

import (
	"context"
)

type bypassKey struct{}

func Bypass(ctx context.Context) context.Context {
	return context.WithValue(ctx, bypassKey{}, true)
}

func IsBypassed(ctx context.Context) bool {
	value := ctx.Value(bypassKey{})
	if value == nil {
		return false
	}

	boolValue, _ := value.(bool)

	return boolValue
}
