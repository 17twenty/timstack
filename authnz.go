package main

import (
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// CreateHashFromPassword - creates hash
func CreateHashFromPassword(password string) string {
	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		log.Println(err)
		return ""
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

// HashMatchesPlaintext - validates if hash matches
func HashMatchesPlaintext(hashedPassword string, plainPassword string) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPassword)
	bytePassword := []byte(plainPassword)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePassword)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

// MaskPrefix - mask a prefix only exposing characters at the end of the string
func MaskPrefix(s string, exposed int) string {
	if exposed >= len(s) || exposed <= 0 {
		return strings.Repeat("*", len(s))
	}
	return strings.Repeat("*", len(s)-exposed) + s[len(s)-exposed:]
}
