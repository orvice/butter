//go:build !darwin

package credentials

// NewSystem returns a store that fails with ErrUnsupported: the first
// phase is macOS only, and a platform without a credential store must not
// fall back to a file.
func NewSystem(string) Store { return Unsupported{} }
