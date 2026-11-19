//go:build darwin

package credentials

import (
	"errors"

	"github.com/egoist/mygo"
)

// keychain keeps tokens in the login Keychain through mygo.Secrets.
type keychain struct {
	service string
}

// NewSystem returns the system credential store for service, the Keychain
// on macOS. On a system without one the returned store fails with
// ErrUnsupported.
func NewSystem(service string) Store {
	if service == "" {
		return keychain{}
	}
	return keychain{service: service}
}

func (k keychain) module() mygo.SecretsModule {
	if k.service == "" {
		return mygo.Secrets
	}
	return mygo.Secrets.Service(k.service)
}

// Get implements Store.
func (k keychain) Get(key string) ([]byte, error) {
	v, err := k.module().Get(key)
	if err != nil {
		return nil, mapErr(err)
	}
	return v, nil
}

// Set implements Store.
func (k keychain) Set(key string, value []byte) error {
	if err := k.module().Set(key, value); err != nil {
		return mapErr(err)
	}
	return nil
}

// Delete implements Store.
func (k keychain) Delete(key string) error {
	if err := k.module().Delete(key); err != nil {
		return mapErr(err)
	}
	return nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, mygo.ErrSecretNotFound):
		return ErrNotFound
	case errors.Is(err, errors.ErrUnsupported):
		return ErrUnsupported
	default:
		return err
	}
}
