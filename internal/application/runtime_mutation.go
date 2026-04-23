package application

import (
	"fmt"

	"github.com/twitchtv/twirp"
)

func mutateWithRuntime[T any](apply func() (T, error), reload func() error, rollback func() error) (T, error) {
	var zero T

	value, err := apply()
	if err != nil {
		return zero, err
	}

	if err := reload(); err != nil {
		if rollback != nil {
			if rollbackErr := rollback(); rollbackErr != nil {
				return zero, twirp.InternalErrorWith(fmt.Errorf("reload runtime: %w; rollback failed: %v", err, rollbackErr))
			}
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
		if rollback != nil {
			if rollbackErr := rollback(); rollbackErr != nil {
				return twirp.InternalErrorWith(fmt.Errorf("reload runtime: %w; rollback failed: %v", err, rollbackErr))
			}
		}
		return err
	}

	return nil
}
