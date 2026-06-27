package application

import "context"

type ConfigRuntime interface {
	ReloadRunner(ctx context.Context) error
	ReloadChannels(ctx context.Context) error
}
