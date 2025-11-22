package service

import "crypto/rand"

// RandReader is a cryptographically secure random number generator.
// It is an alias to crypto/rand.Reader for convenience and testability.
var RandReader = rand.Reader
