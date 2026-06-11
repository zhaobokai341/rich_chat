package service

import (
	"bytes"
	"testing"
)

func TestGenerateRSAKeyPair(t *testing.T) {
	service := NewE2EEncryptionService()

	publicKeyPEM, privateKeyPEM, err := service.GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	if publicKeyPEM == "" {
		t.Error("Public key PEM should not be empty")
	}

	if privateKeyPEM == "" {
		t.Error("Private key PEM should not be empty")
	}

	// Verify the keys can be parsed back
	publicKey, err := service.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	if publicKey == nil {
		t.Fatal("Parsed public key should not be nil")
	}

	privateKey, err := service.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	if privateKey == nil {
		t.Fatal("Parsed private key should not be nil")
	}

	// Verify key size is 2048 bits
	if privateKey.N.BitLen() != 2048 {
		t.Errorf("Expected 2048-bit key, got %d bits", privateKey.N.BitLen())
	}
}

func TestEncryptDecryptMessage(t *testing.T) {
	service := NewE2EEncryptionService()

	// Generate key pair
	publicKeyPEM, privateKeyPEM, err := service.GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	// Test message
	originalMessage := []byte("Hello, this is a secret message with E2EE!")

	// Encrypt message
	encryptedSessionKey, encryptedContent, iv, authTag, err := service.EncryptMessage(originalMessage, publicKeyPEM)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Verify encrypted data is different from original
	if bytes.Equal(encryptedContent, originalMessage) {
		t.Error("Encrypted content should be different from original")
	}

	// Decrypt message
	decrypted, err := service.DecryptMessage(encryptedSessionKey, encryptedContent, iv, authTag, privateKeyPEM)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Verify decrypted message matches original
	if !bytes.Equal(decrypted, originalMessage) {
		t.Errorf("Decrypted message doesn't match original.\nExpected: %s\nGot: %s",
			string(originalMessage), string(decrypted))
	}
}

func TestDecryptWithWrongPrivateKey(t *testing.T) {
	service := NewE2EEncryptionService()

	// Generate two different key pairs
	publicKeyPEM1, _, err := service.GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate first key pair: %v", err)
	}

	_, privateKeyPEM2, err := service.GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate second key pair: %v", err)
	}

	// Encrypt with first public key
	message := []byte("Secret message")
	encryptedSessionKey, encryptedContent, iv, authTag, err := service.EncryptMessage(message, publicKeyPEM1)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Try to decrypt with wrong private key
	_, err = service.DecryptMessage(encryptedSessionKey, encryptedContent, iv, authTag, privateKeyPEM2)
	if err == nil {
		t.Error("Decryption should fail with wrong private key")
	}
}

func TestDecryptWithTamperedData(t *testing.T) {
	service := NewE2EEncryptionService()

	// Generate key pair
	publicKeyPEM, privateKeyPEM, err := service.GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	// Encrypt message
	message := []byte("Secret message")
	encryptedSessionKey, encryptedContent, iv, authTag, err := service.EncryptMessage(message, publicKeyPEM)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Tamper with encrypted content
	encryptedContent[0] ^= 0xFF

	// Try to decrypt tampered data
	_, err = service.DecryptMessage(encryptedSessionKey, encryptedContent, iv, authTag, privateKeyPEM)
	if err == nil {
		t.Error("Decryption should fail with tampered data")
	}
}

func TestEncryptDecryptWithAES(t *testing.T) {
	service := NewE2EEncryptionService()

	// Generate a 32-byte key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	// Test message
	originalMessage := []byte("Test message for AES-256-GCM encryption")

	// Encrypt
	encryptedContent, iv, authTag, err := service.EncryptWithAES(originalMessage, key)
	if err != nil {
		t.Fatalf("AES encryption failed: %v", err)
	}

	// Verify encrypted data is different
	if bytes.Equal(encryptedContent, originalMessage) {
		t.Error("Encrypted content should be different from original")
	}

	// Decrypt
	decrypted, err := service.DecryptWithAES(encryptedContent, iv, authTag, key)
	if err != nil {
		t.Fatalf("AES decryption failed: %v", err)
	}

	// Verify
	if !bytes.Equal(decrypted, originalMessage) {
		t.Errorf("Decrypted message doesn't match original.\nExpected: %s\nGot: %s",
			string(originalMessage), string(decrypted))
	}
}

func TestAESWithWrongKey(t *testing.T) {
	service := NewE2EEncryptionService()

	// Generate two different keys
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	for i := range key1 {
		key1[i] = byte(i)
		key2[i] = byte(i + 1)
	}

	// Encrypt with key1
	message := []byte("Secret message")
	encryptedContent, iv, authTag, err := service.EncryptWithAES(message, key1)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Try to decrypt with key2
	_, err = service.DecryptWithAES(encryptedContent, iv, authTag, key2)
	if err == nil {
		t.Error("Decryption should fail with wrong key")
	}
}

func TestAESWithInvalidKeySize(t *testing.T) {
	service := NewE2EEncryptionService()

	// Test with wrong key size
	wrongKey := make([]byte, 16) // Should be 32 bytes
	message := []byte("Test message")

	_, _, _, err := service.EncryptWithAES(message, wrongKey)
	if err == nil {
		t.Error("Encryption should fail with wrong key size")
	}

	_, err = service.DecryptWithAES([]byte{}, []byte{}, []byte{}, wrongKey)
	if err == nil {
		t.Error("Decryption should fail with wrong key size")
	}
}

func TestSessionKeyEncryptionDecryption(t *testing.T) {
	service := NewE2EEncryptionService()

	// Generate key pair
	publicKeyPEM, privateKeyPEM, err := service.GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	// Generate a session key
	sessionKey := make([]byte, 32)
	for i := range sessionKey {
		sessionKey[i] = byte(i)
	}

	// Encrypt session key
	encryptedSessionKey, err := service.EncryptSessionKey(sessionKey, publicKeyPEM)
	if err != nil {
		t.Fatalf("Session key encryption failed: %v", err)
	}

	// Verify encrypted key is different
	if bytes.Equal(encryptedSessionKey, sessionKey) {
		t.Error("Encrypted session key should be different from original")
	}

	// Decrypt session key
	decryptedKey, err := service.DecryptSessionKey(encryptedSessionKey, privateKeyPEM)
	if err != nil {
		t.Fatalf("Session key decryption failed: %v", err)
	}

	// Verify
	if !bytes.Equal(decryptedKey, sessionKey) {
		t.Error("Decrypted session key doesn't match original")
	}
}

func TestParseInvalidPublicKey(t *testing.T) {
	service := NewE2EEncryptionService()

	// Test with invalid PEM
	_, err := service.ParseRSAPublicKeyFromPEM("invalid pem data")
	if err == nil {
		t.Error("Should fail to parse invalid PEM")
	}

	// Test with valid PEM but not a public key
	privateKeyPEM := `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA2a2rwplBQL1mLbNEZ7Hn5VwO6v2M5V8nV8nV8nV8nV8nV8nV
-----END RSA PRIVATE KEY-----`
	_, err = service.ParseRSAPublicKeyFromPEM(privateKeyPEM)
	if err == nil {
		t.Error("Should fail to parse private key as public key")
	}
}

func TestParseInvalidPrivateKey(t *testing.T) {
	service := NewE2EEncryptionService()

	// Test with invalid PEM
	_, err := service.ParseRSAPrivateKeyFromPEM("invalid pem data")
	if err == nil {
		t.Error("Should fail to parse invalid PEM")
	}

	// Test with valid PEM but not a private key
	publicKeyPEM := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2a2rwplBQL1mLbNEZ7Hn
-----END PUBLIC KEY-----`
	_, err = service.ParseRSAPrivateKeyFromPEM(publicKeyPEM)
	if err == nil {
		t.Error("Should fail to parse public key as private key")
	}
}

func TestMultipleEncryptDecryptCycles(t *testing.T) {
	service := NewE2EEncryptionService()

	// Generate key pair
	publicKeyPEM, privateKeyPEM, err := service.GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	// Test multiple messages
	messages := []string{
		"First message",
		"Second message with more content",
		"Third message: 1234567890!@#$%^&*()",
		"",  // Empty message
		"A", // Single character
	}

	for i, msg := range messages {
		originalMessage := []byte(msg)

		// Encrypt
		encryptedSessionKey, encryptedContent, iv, authTag, err := service.EncryptMessage(originalMessage, publicKeyPEM)
		if err != nil {
			t.Fatalf("Encryption failed for message %d: %v", i+1, err)
		}

		// Decrypt
		decrypted, err := service.DecryptMessage(encryptedSessionKey, encryptedContent, iv, authTag, privateKeyPEM)
		if err != nil {
			t.Fatalf("Decryption failed for message %d: %v", i+1, err)
		}

		// Verify
		if !bytes.Equal(decrypted, originalMessage) {
			t.Errorf("Decrypted message %d doesn't match original.\nExpected: %s\nGot: %s",
				i+1, string(originalMessage), string(decrypted))
		}
	}
}

func TestLargeMessageEncryption(t *testing.T) {
	service := NewE2EEncryptionService()

	// Generate key pair
	publicKeyPEM, privateKeyPEM, err := service.GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	// Create a large message (10KB)
	largeMessage := make([]byte, 10240)
	for i := range largeMessage {
		largeMessage[i] = byte(i % 256)
	}

	// Encrypt
	encryptedSessionKey, encryptedContent, iv, authTag, err := service.EncryptMessage(largeMessage, publicKeyPEM)
	if err != nil {
		t.Fatalf("Large message encryption failed: %v", err)
	}

	// Decrypt
	decrypted, err := service.DecryptMessage(encryptedSessionKey, encryptedContent, iv, authTag, privateKeyPEM)
	if err != nil {
		t.Fatalf("Large message decryption failed: %v", err)
	}

	// Verify
	if !bytes.Equal(decrypted, largeMessage) {
		t.Error("Decrypted large message doesn't match original")
	}
}

func TestKeyPairUniqueness(t *testing.T) {
	service := NewE2EEncryptionService()

	// Generate multiple key pairs and verify they're different
	publicKey1, privateKey1, err := service.GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate first key pair: %v", err)
	}

	publicKey2, privateKey2, err := service.GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate second key pair: %v", err)
	}

	if publicKey1 == publicKey2 {
		t.Error("Different key pairs should have different public keys")
	}

	if privateKey1 == privateKey2 {
		t.Error("Different key pairs should have different private keys")
	}
}
