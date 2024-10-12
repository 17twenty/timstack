package middleware

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		name      string
		plaintext string
		key       string
	}{
		{
			name:      "Basic encryption and decryption",
			plaintext: "This is a secret message",
			key:       "my-secret-key",
		},
		{
			name:      "Empty plaintext",
			plaintext: "",
			key:       "another-secret-key",
		},
		{
			name:      "Long plaintext",
			plaintext: "This is a very long message that spans multiple AES blocks to ensure that the padding and block mode are working correctly across block boundaries.",
			key:       "yet-another-secret-key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt the plaintext
			ciphertext, err := Encrypt(tc.plaintext, tc.key)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Decrypt the ciphertext
			decryptedText, err := Decrypt(ciphertext, tc.key)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Verify that the decrypted text matches the original plaintext
			if decryptedText != tc.plaintext {
				t.Errorf("Expected decrypted text to be %q, but got %q", tc.plaintext, decryptedText)
			}
		})
	}
}

func TestEncryptionDeterminism(t *testing.T) {
	plaintext := "This is a secret message"
	key := "my-secret-key"

	// Encrypt the same plaintext twice
	ciphertext1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	ciphertext2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	// Verify that the two ciphertexts are different (due to random IV)
	if ciphertext1 == ciphertext2 {
		t.Error("Expected different ciphertexts for the same plaintext due to random IV, but got identical ciphertexts")
	}
}

func TestDecryptionWithWrongKey(t *testing.T) {
	plaintext := "This is a secret message"
	correctKey := "correct-secret-key"
	wrongKey := "wrong-secret-key"

	// Encrypt with the correct key
	ciphertext, err := Encrypt(plaintext, correctKey)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Attempt to decrypt with the wrong key
	_, err = Decrypt(ciphertext, wrongKey)
	if err == nil {
		t.Error("Expected decryption with wrong key to fail, but it succeeded")
	}
}

func TestPaddingFunctions(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		blockSize int
	}{
		{"Empty input", []byte{}, 16},
		{"Input smaller than block size", []byte{1, 2, 3}, 16},
		{"Input equal to block size", []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, 16},
		{"Input larger than block size", []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, 16},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			padded := PKCS7Padding(tc.input, tc.blockSize)
			if len(padded)%tc.blockSize != 0 {
				t.Errorf("PKCS7Padding failed: padded length %d is not a multiple of block size %d", len(padded), tc.blockSize)
			}

			unpadded, err := PKCS7UnPadding(padded)
			if err != nil {
				t.Errorf("PKCS7UnPadding failed: %v", err)
			}

			if string(unpadded) != string(tc.input) {
				t.Errorf("PKCS7UnPadding failed: expected %v, got %v", tc.input, unpadded)
			}
		})
	}
}
