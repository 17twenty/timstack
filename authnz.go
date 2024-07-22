package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Recommended key length
	keyLen = 32
)

// CreateArgon2HashFromPassword uses Argon2ID to create and report a DB friendly string
// The draft RFC recommends[2] time=3, and memory=32*1024 is a sensible number.
// If using that amount of memory (32 MB) is not possible in some contexts then the
// time parameter can be increased to compensate.
func CreateArgon2HashFromPassword(password string) string {

	// Stolen from go.dev RFC
	const (
		time     = 3
		mem      = 32 * 1024
		pthreads = 1
	)

	// "Random" salt
	salt := make([]byte, keyLen)
	_, err := rand.Read(salt)
	if err != nil {
		log.Println("RNG Broken")
		panic("RNG failed - OS is not OK")
	}

	// Generate hashed key
	hashKey := argon2.IDKey([]byte(password), salt, time, mem, pthreads, keyLen)

	// Make it match the spec (something about this being already in bcrypt...)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hashKey)

	// Return a string using the standard encoded hash representation.
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, mem, time, pthreads, b64Salt, b64Hash)
}

// Argon2HashMatchesPlaintext is used to match argon2ID for secure passwords
// We compare of the form $argon2id$v=19$m=16,t=2,p=1$dzR5bVM4U2VmWDdvT1J5cQ$b79Ih2lTBIdLb1XfwA2DkA
func Argon2HashMatchesPlaintext(hashedPassword string, plainPassword string) bool {

	vals := strings.Split(hashedPassword, "$")
	if len(vals) != 6 && vals[1] != fmt.Sprintf("v=%d", argon2.Version) {
		log.Println("Bad hash components expected= 6, got", len(vals))
		return false
	}

	params := strings.Split(vals[3], ",")
	if len(params) != 3 {
		log.Println("Bad parameter data expected= 3, got", len(params))
		return false
	}

	// Exract mem, threads etc
	var mem, time, pthreads int
	// Mem
	var err error
	mem, err = strconv.Atoi(strings.ReplaceAll(params[0], "m=", ""))
	if err != nil {
		log.Println("mem parse", err)
		return false
	}
	// Time
	time, err = strconv.Atoi(strings.ReplaceAll(params[1], "t=", ""))
	if err != nil {
		log.Println("time parse", err)
		return false
	}
	// pThreads
	pthreads, err = strconv.Atoi(strings.ReplaceAll(params[2], "p=", ""))
	if err != nil {
		log.Println("pthreads parse", err)
		return false
	}

	// Generate hash as before using extracted values
	var baseErr error
	salt, baseErr := base64.RawStdEncoding.DecodeString(vals[len(vals)-2])
	if baseErr != nil {
		log.Println("salt err", baseErr)
		return false
	}
	hashedKey, baseErr := base64.RawStdEncoding.DecodeString(vals[len(vals)-1])
	if baseErr != nil {
		log.Println("hashedKey err", baseErr)
		return false
	}

	generatedHash := argon2.IDKey([]byte(plainPassword), salt, uint32(time), uint32(mem), uint8(pthreads), keyLen)

	// Compare our new hash with the provided one
	return bytes.Equal(hashedKey, generatedHash)
}

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
