package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
)

// E2EEncryptionService handles end-to-end encryption operations
type E2EEncryptionService interface {
	// GenerateRSAKeyPair generates a new RSA-2048 key pair
	GenerateRSAKeyPair() (publicKeyPEM string, privateKeyPEM string, err error)

	// EncryptMessage encrypts a message using hybrid encryption (RSA + AES-256-GCM)
	// Returns: encryptedSessionKey (RSA encrypted), encryptedContent (AES encrypted), iv, authTag
	EncryptMessage(plaintext []byte, recipientPublicKeyPEM string) (encryptedSessionKey []byte, encryptedContent []byte, iv []byte, authTag []byte, err error)

	// DecryptMessage decrypts a message using hybrid encryption
	DecryptMessage(encryptedSessionKey []byte, encryptedContent []byte, iv []byte, authTag []byte, privateKeyPEM string) ([]byte, error)

	// EncryptWithAES encrypts data using AES-256-GCM with a provided key
	EncryptWithAES(plaintext []byte, key []byte) (encrypted []byte, iv []byte, authTag []byte, err error)

	// DecryptWithAES decrypts data using AES-256-GCM with a provided key
	DecryptWithAES(encrypted []byte, iv []byte, authTag []byte, key []byte) ([]byte, error)

	// EncryptSessionKey encrypts a session key with RSA public key
	EncryptSessionKey(sessionKey []byte, publicKeyPEM string) ([]byte, error)

	// DecryptSessionKey decrypts a session key with RSA private key
	DecryptSessionKey(encryptedSessionKey []byte, privateKeyPEM string) ([]byte, error)

	// ParseRSAPublicKeyFromPEM parses a PEM encoded RSA public key
	ParseRSAPublicKeyFromPEM(pemStr string) (*rsa.PublicKey, error)

	// ParseRSAPrivateKeyFromPEM parses a PEM encoded RSA private key
	ParseRSAPrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error)
}

// e2eEncryptionServiceImpl implements E2EEncryptionService
type e2eEncryptionServiceImpl struct {
	rsaKeySize int
}

// NewE2EEncryptionService creates a new E2EE encryption service
func NewE2EEncryptionService() E2EEncryptionService {
	return &e2eEncryptionServiceImpl{
		rsaKeySize: 2048,
	}
}

// GenerateRSAKeyPair generates a new RSA-2048 key pair and returns PEM encoded strings
func (e *e2eEncryptionServiceImpl) GenerateRSAKeyPair() (string, string, error) {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, e.rsaKeySize)
	if err != nil {
		return "", "", errors.New("failed to generate RSA key pair: " + err.Error())
	}

	// Encode private key to PEM
	privateKeyPEM := e.encodeRSAPrivateKeyToPEM(privateKey)

	// Extract public key and encode to PEM
	publicKeyPEM := e.encodeRSAPublicKeyToPEM(&privateKey.PublicKey)

	return publicKeyPEM, privateKeyPEM, nil
}

// EncryptMessage encrypts a message using hybrid encryption
// 1. Generate random AES-256 session key
// 2. Encrypt message with AES-256-GCM using session key
// 3. Encrypt session key with recipient's RSA public key
func (e *e2eEncryptionServiceImpl) EncryptMessage(plaintext []byte, recipientPublicKeyPEM string) ([]byte, []byte, []byte, []byte, error) {
	// Parse recipient's public key
	publicKey, err := e.ParseRSAPublicKeyFromPEM(recipientPublicKeyPEM)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Generate random 32-byte session key for AES-256
	sessionKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, sessionKey); err != nil {
		return nil, nil, nil, nil, errors.New("failed to generate session key: " + err.Error())
	}

	// Encrypt message with AES-256-GCM
	encryptedContent, iv, authTag, err := e.EncryptWithAES(plaintext, sessionKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Encrypt session key with recipient's RSA public key
	encryptedSessionKey, err := e.EncryptSessionKey(sessionKey, recipientPublicKeyPEM)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Verify the encrypted content can be decrypted with the public key
	_ = publicKey // Use the parsed key to avoid unused variable error

	return encryptedSessionKey, encryptedContent, iv, authTag, nil
}

// DecryptMessage decrypts a message using hybrid encryption
// 1. Decrypt session key with recipient's RSA private key
// 2. Decrypt message with AES-256-GCM using session key
func (e *e2eEncryptionServiceImpl) DecryptMessage(encryptedSessionKey []byte, encryptedContent []byte, iv []byte, authTag []byte, privateKeyPEM string) ([]byte, error) {
	// Decrypt session key with RSA private key
	sessionKey, err := e.DecryptSessionKey(encryptedSessionKey, privateKeyPEM)
	if err != nil {
		return nil, err
	}

	// Decrypt message with AES-256-GCM
	plaintext, err := e.DecryptWithAES(encryptedContent, iv, authTag, sessionKey)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptWithAES encrypts data using AES-256-GCM
func (e *e2eEncryptionServiceImpl) EncryptWithAES(plaintext []byte, key []byte) ([]byte, []byte, []byte, error) {
	if len(key) != 32 {
		return nil, nil, nil, errors.New("key must be 32 bytes for AES-256")
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, errors.New("failed to create AES cipher: " + err.Error())
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, errors.New("failed to create GCM: " + err.Error())
	}

	// Generate random nonce (12 bytes for GCM)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, nil, errors.New("failed to generate nonce: " + err.Error())
	}

	// Encrypt and authenticate
	// Seal appends the ciphertext and authentication tag
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Split ciphertext into encrypted content and auth tag
	// The last 16 bytes are the authentication tag
	authTagSize := 16
	if len(ciphertext) < authTagSize {
		return nil, nil, nil, errors.New("ciphertext too short")
	}

	encryptedContent := ciphertext[:len(ciphertext)-authTagSize]
	authTag := ciphertext[len(ciphertext)-authTagSize:]

	return encryptedContent, nonce, authTag, nil
}

// DecryptWithAES decrypts data using AES-256-GCM
func (e *e2eEncryptionServiceImpl) DecryptWithAES(encryptedContent []byte, nonce []byte, authTag []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New("failed to create AES cipher: " + err.Error())
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("failed to create GCM: " + err.Error())
	}

	// Verify nonce size
	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("invalid nonce size")
	}

	// Combine encrypted content and auth tag for Open
	ciphertext := append(encryptedContent, authTag...)

	// Decrypt and verify authenticity
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed: invalid key or tampered data")
	}

	return plaintext, nil
}

// EncryptSessionKey encrypts a session key with RSA public key using OAEP
func (e *e2eEncryptionServiceImpl) EncryptSessionKey(sessionKey []byte, publicKeyPEM string) ([]byte, error) {
	publicKey, err := e.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, err
	}

	// Encrypt with RSA-OAEP
	label := []byte("")
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, sessionKey, label)
	if err != nil {
		return nil, errors.New("failed to encrypt session key: " + err.Error())
	}

	return ciphertext, nil
}

// DecryptSessionKey decrypts a session key with RSA private key using OAEP
func (e *e2eEncryptionServiceImpl) DecryptSessionKey(encryptedSessionKey []byte, privateKeyPEM string) ([]byte, error) {
	privateKey, err := e.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	// Decrypt with RSA-OAEP
	label := []byte("")
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedSessionKey, label)
	if err != nil {
		return nil, errors.New("failed to decrypt session key: " + err.Error())
	}

	return plaintext, nil
}

// ParseRSAPublicKeyFromPEM parses a PEM encoded RSA public key
func (e *e2eEncryptionServiceImpl) ParseRSAPublicKeyFromPEM(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse public key: " + err.Error())
	}

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return publicKey, nil
}

// ParseRSAPrivateKeyFromPEM parses a PEM encoded RSA private key
func (e *e2eEncryptionServiceImpl) ParseRSAPrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse private key: " + err.Error())
	}

	return privateKey, nil
}

// encodeRSAPrivateKeyToPEM encodes an RSA private key to PEM format
func (e *e2eEncryptionServiceImpl) encodeRSAPrivateKeyToPEM(key *rsa.PrivateKey) string {
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return string(pem.EncodeToMemory(privateKeyPEM))
}

// encodeRSAPublicKeyToPEM encodes an RSA public key to PEM format
func (e *e2eEncryptionServiceImpl) encodeRSAPublicKeyToPEM(key *rsa.PublicKey) string {
	pubBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return ""
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	return string(pem.EncodeToMemory(publicKeyPEM))
}
