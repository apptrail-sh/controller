package hooks

import (
	"context"
	"github.com/apptrail-sh/controller/internal/model"
)

type Notifier interface {
	Notify(ctx context.Context, update model.WorkloadUpdate) error
}
