package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"log"
	"strings"

	"golang.org/x/crypto/argon2"
)

func getDefaultArgon2Params() *Argon2Params {
	config, err := GetConfig()
	if err != nil {
		log.Fatalf("[Error] Failed to load config: %v", err)
	}
	params := &Argon2Params{
		Memory:      config.Password.Memory * 1024,
		Iterations:  config.Password.Iterations,
		Parallelism: config.Password.Parallelism,
		SaltLength:  config.Password.SaltLength,
		KeyLength:   config.Password.KeyLength,
	}
	return params
}

// HashBytes returns the hex string of the given hasher after writing data.
func HashBytes(h hash.Hash, data []byte) string {
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// HashPassword uses Argon2id to securely hash the original password
func HashPassword(password string) (string, error) {
	params := getDefaultArgon2Params()

	// Generate random salt
	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Use Argon2id to compute hash
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Encode as string to store (compatible with PHC format)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=2$c2FsdA$hash...
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Iterations,
		params.Parallelism,
		b64Salt,
		b64Hash,
	), nil
}

// CheckPassword Verify if the original password matches the Argon2id hash
func CheckPassword(password, hash string) bool {
	ok, err := verifyPassword(password, hash)
	return err == nil && ok
}

// verifyPassword It is an internal validation function that returns Boolean values and errors
func verifyPassword(password, hash string) (bool, error) {
	parts := parseHash(hash)
	if parts == nil {
		return false, errors.New("invalid hash format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts.SaltBase64)
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts.HashBase64)
	if err != nil {
		return false, err
	}

	// Recalculate Hash
	computed := argon2.IDKey(
		[]byte(password),
		salt,
		parts.Iterations,
		parts.Memory,
		parts.Parallelism,
		uint32(len(expectedHash)),
	)

	// Constant time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(computed, expectedHash) == 1, nil
}

// parseHash Parse the $argon2id$... format string
func parseHash(hash string) *HashParts {
	vals := strings.Split(hash, "$")
	if len(vals) != 6 {
		return nil
	}

	var m, t uint32
	var p uint8
	_, err := fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return nil
	}

	return &HashParts{
		Memory:      m,
		Iterations:  t,
		Parallelism: p,
		SaltBase64:  vals[4],
		HashBase64:  vals[5],
	}
}
