package application

import (
	"errors"

	"connectrpc.com/connect"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/transport/connectx"
)

func toConnectError(err error) *connect.Error {
	var cerr *connect.Error
	if errors.As(err, &cerr) {
		return cerr
	}
	if errors.Is(err, configrepo.ErrNotFound) {
		return connectx.NotFound(err.Error())
	}
	if errors.Is(err, configrepo.ErrAlreadyExists) {
		return connect.NewError(connect.CodeAlreadyExists, errors.New(err.Error()))
	}
	return connectx.InternalWith(err)
}
