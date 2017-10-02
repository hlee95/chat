package chatauth

import (
  crand "crypto/rand"
  "math/rand"
  "golang.org/x/crypto/bcrypt"
)
// This file provides helper functions for handling user authentication.
//
// (For now, just provides helpers for randomly generating salts,
//  hashing passwords, comparing hashes. In the future, it would
//  be pulled into its own package and provide functions for logging in,
//  creating and invalidating session tokens, etc.)

const HASH_COST = 14

// Returns a random string of the given length in bytes.
func generateRandomBytes(numBytes int) []byte {
  bytes := make([]byte, numBytes)
  crand.Read(bytes)
  return bytes
}

// Return a random integer.
func GenerateID() int {
  return rand.Int()
}

// Return a salt of the desired length.
func GenerateSalt(numBytes int) []byte {
  return generateRandomBytes(numBytes)
}

// Given a salt and a password, generate the hash.
func HashPasswordWithSalt(password string, salt []byte) ([]byte, error) {
  hash, err := bcrypt.GenerateFromPassword(
    append([]byte(password), salt...),
    HASH_COST)
  return hash, err
}

// Given a password, salt and hash, return token if correct, otherwise error.
// This token can be used to validate future requests from this user for the
// remainder of their session (although that is not implemented in this project).
func Authenticate(password string, salt []byte, hash []byte) ([]byte, error) {
  if err := bcrypt.CompareHashAndPassword(hash, append([]byte(password), salt...)); err != nil {
    return nil, err
  }
  // Password is good, return a token (256 bits).
  return generateRandomBytes(32), nil
}
