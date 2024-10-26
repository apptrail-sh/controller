package hooks

import (
	"context"
	"github.com/apptrail-sh/controller/internal/model"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type NotifierQueue struct {
	UpdateChan <-chan model.WorkloadUpdate
	notifiers  []Notifier
}

func NewNotifierQueue(updateChan <-chan model.WorkloadUpdate, notifiers []Notifier) *NotifierQueue {
	return &NotifierQueue{
		UpdateChan: updateChan,
		notifiers:  notifiers,
	}
}

func (nq *NotifierQueue) Loop() {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	for {
		select {
		case update := <-nq.UpdateChan:
			// Notify all registered notifiers
			for _, notifier := range nq.notifiers {
				if update.PreviousVersion != "" {
					err := notifier.Notify(ctx, update)
					if err != nil {
						logger.Error(err, "failed to notify")
					}
				}
			}
		}
	}
}
