package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func VerifyPassword(password, encodeHash string) error {
	parts := strings.Split(encodeHash, ".")
	if len(parts) != 2 {
		return ErrorHandler(errors.New("Invalid encoded hash format"), "Invalid encoded hash format")
		//http.Error(w, "Invalid encoded hash format", http.StatusForbidden)
		//return
	}

	saltBase64 := parts[0]
	hashedPasswordBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return ErrorHandler(err, "failed to decode the salt")
		//http.Error(w, "failed to decode the salt", http.StatusForbidden)
		//return
	}

	hashedPassword, err := base64.StdEncoding.DecodeString(hashedPasswordBase64)
	if err != nil {
		return ErrorHandler(err, "failed to decode the hashed password")
		//http.Error(w, "failed to decode the hashed password", http.StatusForbidden)
		//return
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	if len(hash) != len(hashedPassword) {
		return ErrorHandler(errors.New("incorrect password"), "incorrect password")
		//http.Error(w, "incorrect password", http.StatusForbidden)
		//return
	}

	if subtle.ConstantTimeCompare(hash, hashedPassword) == 1 {
		return nil
	}

	return ErrorHandler(errors.New("incorrect password"), "incorrect password")
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrorHandler(errors.New("password is blank"), "please enter password")
	}
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", ErrorHandler(errors.New("failed to generate salt"), "error adding data")
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)
	return encodedHash, nil
}
