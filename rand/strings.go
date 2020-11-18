package rand

import (
	"crypto/rand"
	"encoding/base64"
)

const rememberTokenBytes = 32

// Bytes will help us generate n random bytes or will
// return an error if there was one.
func Bytes(n int) ([]byte, error) {
	// create a byte slice of size n
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// String ...
func String(nBytes int) (string, error) {
	b, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// RememberToken ...
func RememberToken() (string, error) {
	return String(rememberTokenBytes)
}
