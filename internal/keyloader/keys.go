package keyloader

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"go-jwks-server/internal/keyfiles"
	"os"
	"path/filepath"
	"strings"

	"github.com/lestrrat-go/jwx/jwk"
)

func LoadPublicKey(key []byte) (jwk.Key, error) {
	keyPem, _ := pem.Decode(key)
	if keyPem == nil {
		return nil, errors.New("failed to decode PEM file")
	}

	if keyPem.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("public key wrong type: %s", keyPem.Type)
	}

	parsedKey, err := x509.ParsePKIXPublicKey(keyPem.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing public key: %w", err)
	}

	jwkPubKey, err := jwk.New(parsedKey)
	if err != nil {
		return nil, fmt.Errorf("creating JWK: %w", err)
	}

	return jwkPubKey, nil
}

func LoadPublicKeyFromFile(file string) (jwk.Key, error) {
	pubBuf, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading public key file: %w", err)
	}

	return LoadPublicKey(pubBuf)
}

func loadKeys(dir string) (jwk.Set, error) {
	fileMetadata, skipped, err := keyfiles.GetFileMetadata(dir)
	if err != nil {
		return nil, fmt.Errorf("getting file metadata: %w", err)
	}

	keySet := jwk.NewSet()

	loaded := map[string]string{}

	for _, f := range fileMetadata {
		fullPath := filepath.Join(dir, f.Name)

		key, err := LoadPublicKeyFromFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("loading key from %s: %w", fullPath, err)
		}

		keyId := f.Name
		if strings.HasSuffix(strings.ToLower(keyId), ".pub") {
			keyId = keyId[:len(keyId)-4]
		}

		key.Set(jwk.KeyIDKey, keyId)
		key.Set(jwk.KeyUsageKey, jwk.ForSignature)

		added := keySet.Add(key)

		if !added {
			log.Warn().Str("filename", f.Name).Str("keyId", keyId).Msg("key already loaded")
		}

		loaded[f.Name] = keyId
	}

	if len(skipped) > 0 {
		log.Info().Interface("skipped", skipped).Interface("loaded", loaded).Msg("loaded keys")
	}

	return keySet, nil
}
