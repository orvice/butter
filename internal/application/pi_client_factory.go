package application

import (
	"context"
	"fmt"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	"go.orx.me/apps/butter/internal/runtime/pibox"
	"go.orx.me/apps/butter/internal/secretbox"

	internalagent "go.orx.me/apps/butter/internal/agent"
)

// NewPiClientFactory returns a PiClientFactory that decrypts the ButterBox
// token via the secretbox keyring and builds a PiService Connect client.
func NewPiClientFactory(repo butterboxrepo.Repository, keyring *secretbox.Keyring) internalagent.PiClientFactory {
	return func(ctx context.Context, workspaceID, butterboxID string) (pibox.PiClient, error) {
		box, err := repo.Get(ctx, workspaceID, butterboxID)
		if err != nil {
			return nil, fmt.Errorf("butterbox %q: %w", butterboxID, err)
		}
		if !box.GetEnabled() {
			return nil, fmt.Errorf("butterbox %q is disabled", box.GetName())
		}

		token := ""
		if box.GetCredentialSet() {
			if keyring == nil {
				return nil, fmt.Errorf("butterbox %q: credential encryption is not configured", box.GetName())
			}
			cred, err := repo.GetCredential(ctx, workspaceID, butterboxID)
			if err != nil {
				return nil, fmt.Errorf("butterbox %q: get credential: %w", box.GetName(), err)
			}
			plaintext, err := keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
			if err != nil {
				return nil, fmt.Errorf("butterbox %q: decrypt credential: %w", box.GetName(), err)
			}
			token = string(plaintext)
		}

		return pibox.NewPiClient(box.GetBaseUrl(), token), nil
	}
}
