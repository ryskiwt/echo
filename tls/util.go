package main

import "context"

func isDone(ctx context.Context) bool {
	select {
	default:
	case _, ok := <-ctx.Done():
		return !ok
	}
	return false
}
