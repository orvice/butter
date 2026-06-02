package application

import (
	"context"
	"fmt"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/transport/connectx"
)

func mutateWithRuntime[T any](apply func() (T, error), reload func() error, rollback func() error) (T, error) {
	var zero T

	value, err := apply()
	if err != nil {
		return zero, err
	}

	if err := reload(); err != nil {
		log.FromContext(context.Background()).Warn("runtime reload failed after mutation, attempting rollback", "err", err)
		if rollback != nil {
			if rollbackErr := rollback(); rollbackErr != nil {
				log.FromContext(context.Background()).Error("runtime rollback failed",
					"reload_err", err, "rollback_err", rollbackErr)
				return zero, connectx.InternalWith(fmt.Errorf("reload runtime: %w; rollback failed: %v", err, rollbackErr))
			}
			log.FromContext(context.Background()).Info("runtime rollback succeeded after reload failure", "reload_err", err)
		}
		return zero, err
	}

	return value, nil
}

func deleteWithRuntime(apply func() error, reload func() error, rollback func() error) error {
	if err := apply(); err != nil {
		return err
	}

	if err := reload(); err != nil {
		log.FromContext(context.Background()).Warn("runtime reload failed after delete, attempting rollback", "err", err)
		if rollback != nil {
			if rollbackErr := rollback(); rollbackErr != nil {
				log.FromContext(context.Background()).Error("runtime rollback failed after delete",
					"reload_err", err, "rollback_err", rollbackErr)
				return connectx.InternalWith(fmt.Errorf("reload runtime: %w; rollback failed: %v", err, rollbackErr))
			}
			log.FromContext(context.Background()).Info("runtime rollback succeeded after delete reload failure", "reload_err", err)
		}
		return err
	}

	return nil
}
