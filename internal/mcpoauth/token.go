package mcpoauth

import (
	"encoding/json"
	"time"

	"golang.org/x/oauth2"
)

type TokenPayload struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
	Scopes       []string  `json:"scopes,omitempty"`
}

func tokenPayloadFromOAuth(tok *oauth2.Token, fallbackScopes []string) TokenPayload {
	return TokenPayload{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry,
		Scopes:       uniqueStrings(fallbackScopes),
	}
}

func oauthTokenFromPayload(payload TokenPayload) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		TokenType:    payload.TokenType,
		Expiry:       payload.Expiry,
	}
}

func encryptTokenPayload(cipher *Cipher, payload TokenPayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return cipher.Encrypt(data)
}

func decryptTokenPayload(cipher *Cipher, encrypted string) (TokenPayload, error) {
	data, err := cipher.Decrypt(encrypted)
	if err != nil {
		return TokenPayload{}, err
	}
	var payload TokenPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return TokenPayload{}, err
	}
	return payload, nil
}
