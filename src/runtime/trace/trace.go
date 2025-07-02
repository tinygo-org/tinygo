// Stubs for the runtime/trace package
package trace

import (
	"context"
	"errors"
	"io"
)

func Start(w io.Writer) error {
	return errors.New("not implemented")
}

func Stop() {}

func NewTask(pctx context.Context, taskType string) (ctx context.Context, task *Task) {
	return context.TODO(), nil
}

type Task struct{}

func (t *Task) End() {}

func Log(ctx context.Context, category, message string) {}

func Logf(ctx context.Context, category, format string, args ...any) {}

func WithRegion(ctx context.Context, regionType string, fn func()) {
	fn()
}

func StartRegion(ctx context.Context, regionType string) *Region {
	return nil
}

type Region struct{}

func (r *Region) End() {}

func IsEnabled() bool {
	return false
}
