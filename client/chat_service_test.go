package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"hash"
	"testing"
)

// Helper wrappers to avoid import conflicts
func sha256New() hash.Hash {
	return sha256.New()
}

func aesNewCipher(key []byte) (cipher.Block, error) {
	return aes.NewCipher(key)
}

func cipherNewGCM(block cipher.Block) (cipher.AEAD, error) {
	return cipher.NewGCM(block)
}

func base64StdEncodingEncodeToString(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func TestChatService_GenerateAndUploadKeys(t *testing.T) {
	mockAPI := &MockChatAPIClient{}
	mockConfig := NewMockConfigManager()
	lp := NewLanguagePackWrapper("client/main.json", "en")

	service := NewChatService(mockAPI, mockConfig, lp)

	userID := 1

	err := service.GenerateAndUploadKeys(userID, "mock-token", "verify-token")
	if err != nil {
		t.Fatalf("GenerateAndUploadKeys failed: %v", err)
	}

	hasKey, err := mockConfig.HasEncryptionKey(userID)
	if err != nil {
		t.Fatalf("HasEncryptionKey failed: %v", err)
	}
	if !hasKey {
		t.Fatal("Expected encryption key to exist after generation")
	}

	keyData, err := mockConfig.GetEncryptionKey(userID)
	if err != nil {
		t.Fatalf("GetEncryptionKey failed: %v", err)
	}
	if keyData == "" {
		t.Fatal("Expected non-empty key data")
	}

	block, _ := pem.Decode([]byte(keyData))
	if block == nil {
		t.Fatal("Failed to decode PEM block from stored key")
	}
	if block.Type != "RSA PRIVATE KEY" {
		t.Fatalf("Expected PEM type 'RSA PRIVATE KEY', got '%s'", block.Type)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse stored private key: %v", err)
	}
	if privateKey.N.BitLen() != 4096 {
		t.Fatalf("Expected 4096-bit key, got %d bits", privateKey.N.BitLen())
	}
}

func TestChatService_GenerateAndUploadKeys_StoreError(t *testing.T) {
	mockAPI := &MockChatAPIClient{}
	mockConfig := NewMockConfigManager()
	mockConfig.SaveEncryptionKeyFunc = func(userID int, privateKeyPEM string) error {
		return &mockError{"save failed"}
	}
	lp := NewLanguagePackWrapper("client/main.json", "en")

	service := NewChatService(mockAPI, mockConfig, lp)

	err := service.GenerateAndUploadKeys(1, "mock-token", "verify-token")
	if err == nil {
		t.Fatal("Expected error when key save fails")
	}
}

func TestChatService_GenerateAndUploadKeys_UploadError(t *testing.T) {
	mockAPI := &MockChatAPIClient{}
	mockAPI.StoreUserKeyFunc = func(userID int, token, publicKey string, encryptedPrivateKey []byte, keyAlgorithm string, verifyToken string) error {
		return &mockError{"upload failed"}
	}
	mockConfig := NewMockConfigManager()
	lp := NewLanguagePackWrapper("client/main.json", "en")

	service := NewChatService(mockAPI, mockConfig, lp)

	err := service.GenerateAndUploadKeys(1, "mock-token", "verify-token")
	if err == nil {
		t.Fatal("Expected error when key upload fails")
	}
}

func TestChatService_HasEncryptionKey(t *testing.T) {
	mockAPI := &MockChatAPIClient{}
	mockConfig := NewMockConfigManager()
	lp := NewLanguagePackWrapper("client/main.json", "en")

	service := NewChatService(mockAPI, mockConfig, lp)

	hasKey, err := service.HasEncryptionKey(1)
	if err != nil {
		t.Fatalf("HasEncryptionKey failed: %v", err)
	}
	if hasKey {
		t.Fatal("Expected no key before generation")
	}

	err = service.GenerateAndUploadKeys(1, "mock-token", "verify-token")
	if err != nil {
		t.Fatalf("GenerateAndUploadKeys failed: %v", err)
	}

	hasKey, err = service.HasEncryptionKey(1)
	if err != nil {
		t.Fatalf("HasEncryptionKey failed: %v", err)
	}
	if !hasKey {
		t.Fatal("Expected key after generation")
	}
}

func TestChatService_LoadEncryptionKey(t *testing.T) {
	mockAPI := &MockChatAPIClient{}
	mockConfig := NewMockConfigManager()
	lp := NewLanguagePackWrapper("client/main.json", "en")

	service := NewChatService(mockAPI, mockConfig, lp)

	err := service.GenerateAndUploadKeys(1, "mock-token", "verify-token")
	if err != nil {
		t.Fatalf("GenerateAndUploadKeys failed: %v", err)
	}

	err = service.LoadEncryptionKey(1)
	if err != nil {
		t.Fatalf("LoadEncryptionKey failed: %v", err)
	}

	if service.privateKey == nil {
		t.Fatal("Expected private key to be loaded")
	}
}

func TestChatService_LoadEncryptionKey_NotFound(t *testing.T) {
	mockAPI := &MockChatAPIClient{}
	mockConfig := NewMockConfigManager()
	lp := NewLanguagePackWrapper("client/main.json", "en")

	service := NewChatService(mockAPI, mockConfig, lp)

	err := service.LoadEncryptionKey(999)
	if err == nil {
		t.Fatal("Expected error when key not found")
	}
}

func TestChatService_EncryptDecryptMessage(t *testing.T) {
	senderAPI := &MockChatAPIClient{}
	senderConfig := NewMockConfigManager()
	senderLp := NewLanguagePackWrapper("client/main.json", "en")
	senderService := NewChatService(senderAPI, senderConfig, senderLp)

	recipientAPI := &MockChatAPIClient{}
	recipientConfig := NewMockConfigManager()
	recipientLp := NewLanguagePackWrapper("client/main.json", "en")
	recipientService := NewChatService(recipientAPI, recipientConfig, recipientLp)

	err := senderService.GenerateAndUploadKeys(1, "mock-token", "verify-token")
	if err != nil {
		t.Fatalf("Sender GenerateAndUploadKeys failed: %v", err)
	}

	err = recipientService.GenerateAndUploadKeys(2, "mock-token", "verify-token")
	if err != nil {
		t.Fatalf("Recipient GenerateAndUploadKeys failed: %v", err)
	}

	err = senderService.LoadEncryptionKey(1)
	if err != nil {
		t.Fatalf("Sender LoadEncryptionKey failed: %v", err)
	}

	err = recipientService.LoadEncryptionKey(2)
	if err != nil {
		t.Fatalf("Recipient LoadEncryptionKey failed: %v", err)
	}

	recipientPubKeyBytes, err := x509.MarshalPKIXPublicKey(&recipientService.privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal recipient public key: %v", err)
	}
	recipientPubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: recipientPubKeyBytes,
	})

	if len(recipientPubKeyPEM) == 0 {
		t.Fatal("Expected non-empty PEM encoded public key")
	}

	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		t.Fatalf("Failed to generate session key: %v", err)
	}

	originalMessage := "Hello, this is a test message!"

	encryptedMsg, err := encryptMessageForTest(originalMessage, sessionKey, &recipientService.privateKey.PublicKey)
	if err != nil {
		t.Fatalf("encryptMessageForTest failed: %v", err)
	}

	decrypted, err := recipientService.DecryptMessage(encryptedMsg)
	if err != nil {
		t.Fatalf("DecryptMessage failed: %v", err)
	}

	if decrypted != originalMessage {
		t.Fatalf("Expected decrypted message '%s', got '%s'", originalMessage, decrypted)
	}
}

func TestChatService_DecryptMessage_NoPrivateKey(t *testing.T) {
	mockAPI := &MockChatAPIClient{}
	mockConfig := NewMockConfigManager()
	lp := NewLanguagePackWrapper("client/main.json", "en")

	service := NewChatService(mockAPI, mockConfig, lp)

	msg := &ChatMessage{
		EncryptedSessionKey: "test",
		EncryptedContent:    "test",
		Iv:                  "test",
		AuthTag:             "test",
	}

	_, err := service.DecryptMessage(msg)
	if err == nil {
		t.Fatal("Expected error when private key not loaded")
	}
}

func TestChatService_GetMessages(t *testing.T) {
	mockAPI := &MockChatAPIClient{}
	mockConfig := NewMockConfigManager()
	lp := NewLanguagePackWrapper("client/main.json", "en")

	service := NewChatService(mockAPI, mockConfig, lp)

	ch := service.GetMessages()
	if ch != nil {
		t.Fatal("Expected nil message channel before connect")
	}
}

func TestChatService_Disconnect(t *testing.T) {
	mockAPI := &MockChatAPIClient{}
	mockConfig := NewMockConfigManager()
	lp := NewLanguagePackWrapper("client/main.json", "en")

	service := NewChatService(mockAPI, mockConfig, lp)

	err := service.Disconnect()
	if err != nil {
		t.Fatalf("Disconnect should not error when not connected: %v", err)
	}
}

func TestChatAPIClient_Interface(t *testing.T) {
	mockAPI := &MockChatAPIClient{}

	userInfo, err := mockAPI.GetUserBasicInfo(1, "token", "test-token", 1)
	if err != nil {
		t.Fatalf("GetUserBasicInfo failed: %v", err)
	}
	if userInfo == nil {
		t.Fatal("Expected non-nil user info")
	}

	sessionID, err := mockAPI.CreateChatSession(1, "token", "test-token", 2)
	if err != nil {
		t.Fatalf("CreateChatSession failed: %v", err)
	}
	if sessionID == 0 {
		t.Fatal("Expected non-zero session ID")
	}

	sessions, err := mockAPI.GetUserSessions(1, "token", "test-token")
	if err != nil {
		t.Fatalf("GetUserSessions failed: %v", err)
	}
	if sessions == nil {
		t.Fatal("Expected non-nil sessions slice")
	}
}

func TestChatAPIClient_CreateChatSession_Custom(t *testing.T) {
	mockAPI := &MockChatAPIClient{
		CreateChatSessionFunc: func(userID int, token, verifyToken string, recipientID int) (int, error) {
			if userID != 1 || recipientID != 2 {
				return 0, &mockError{"unexpected user IDs"}
			}
			return 42, nil
		},
	}

	sessionID, err := mockAPI.CreateChatSession(1, "token", "test-token", 2)
	if err != nil {
		t.Fatalf("CreateChatSession failed: %v", err)
	}

	if sessionID != 42 {
		t.Fatalf("Expected session ID 42, got %d", sessionID)
	}
}

func TestChatAPIClient_GetUserSessions_Custom(t *testing.T) {
	expectedSessions := []*ChatSession{
		{SessionID: 1, SessionType: "direct", PartnerID: 2, PartnerName: "bob"},
		{SessionID: 2, SessionType: "direct", PartnerID: 3, PartnerName: "charlie"},
	}

	mockAPI := &MockChatAPIClient{
		GetUserSessionsFunc: func(userID int, token, verifyToken string) ([]*ChatSession, error) {
			if userID != 1 {
				return nil, &mockError{"unexpected user ID"}
			}
			return expectedSessions, nil
		},
	}

	sessions, err := mockAPI.GetUserSessions(1, "token", "test-token")
	if err != nil {
		t.Fatalf("GetUserSessions failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("Expected 2 sessions, got %d", len(sessions))
	}

	if sessions[0].PartnerName != "bob" {
		t.Fatalf("Expected first session partner 'bob', got '%s'", sessions[0].PartnerName)
	}
}

func TestChatAPIClient_GetUserPublicKey_Custom(t *testing.T) {
	expectedKey := "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----"

	mockAPI := &MockChatAPIClient{
		GetUserPublicKeyFunc: func(currentUserID int, token, verifyToken string, targetUserID int) (string, string, error) {
			if targetUserID != 2 {
				return "", "", &mockError{"unexpected user ID"}
			}
			return expectedKey, "RSA-2048", nil
		},
	}

	pubKey, algorithm, err := mockAPI.GetUserPublicKey(1, "token", "test-token", 2)
	if err != nil {
		t.Fatalf("GetUserPublicKey failed: %v", err)
	}

	if pubKey != expectedKey {
		t.Fatalf("Expected public key '%s', got '%s'", expectedKey, pubKey)
	}

	if algorithm != "RSA-2048" {
		t.Fatalf("Expected algorithm 'RSA-2048', got '%s'", algorithm)
	}
}

func TestChatAPIClient_GetUserPublicKey_NotFound(t *testing.T) {
	mockAPI := &MockChatAPIClient{}

	_, _, err := mockAPI.GetUserPublicKey(1, "token", "test-token", 999)
	if err == nil {
		t.Fatal("Expected error when public key not found")
	}
}

func TestChatAPIClient_StoreUserKey_Custom(t *testing.T) {
	var storedKey string
	var storedAlgorithm string

	mockAPI := &MockChatAPIClient{
		StoreUserKeyFunc: func(userID int, token, publicKey string, encryptedPrivateKey []byte, keyAlgorithm string, verifyToken string) error {
			if userID != 1 {
				return &mockError{"unexpected user ID"}
			}
			storedKey = publicKey
			storedAlgorithm = keyAlgorithm
			return nil
		},
	}

	err := mockAPI.StoreUserKey(1, "token", "test-key", nil, "RSA-2048", "test-token")
	if err != nil {
		t.Fatalf("StoreUserKey failed: %v", err)
	}

	if storedKey != "test-key" {
		t.Fatalf("Expected stored key 'test-key', got '%s'", storedKey)
	}

	if storedAlgorithm != "RSA-2048" {
		t.Fatalf("Expected stored algorithm 'RSA-2048', got '%s'", storedAlgorithm)
	}
}

func TestUserBasicInfo_Struct(t *testing.T) {
	user := &UserBasicInfo{
		ID:       1,
		Username: "testuser",
		Nickname: "Test User",
		Bio:      "Test bio",
	}

	if user.ID != 1 {
		t.Fatalf("Expected ID 1, got %d", user.ID)
	}

	if user.Username != "testuser" {
		t.Fatalf("Expected username 'testuser', got '%s'", user.Username)
	}

	if user.Nickname != "Test User" {
		t.Fatalf("Expected nickname 'Test User', got '%s'", user.Nickname)
	}

	if user.Bio != "Test bio" {
		t.Fatalf("Expected bio 'Test bio', got '%s'", user.Bio)
	}
}

func TestChatSession_Struct(t *testing.T) {
	session := &ChatSession{
		SessionID:   1,
		SessionType: "direct",
		PartnerID:   2,
		PartnerName: "bob",
		MemberCount: 2,
	}

	if session.SessionID != 1 {
		t.Fatalf("Expected session ID 1, got %d", session.SessionID)
	}

	if session.SessionType != "direct" {
		t.Fatalf("Expected session type 'direct', got '%s'", session.SessionType)
	}

	if session.PartnerID != 2 {
		t.Fatalf("Expected partner ID 2, got %d", session.PartnerID)
	}

	if session.PartnerName != "bob" {
		t.Fatalf("Expected partner name 'bob', got '%s'", session.PartnerName)
	}

	if session.MemberCount != 2 {
		t.Fatalf("Expected member count 2, got %d", session.MemberCount)
	}
}

func TestChatMessage_Struct(t *testing.T) {
	msg := &ChatMessage{
		MessageID:           1,
		SenderID:            2,
		SessionID:           3,
		EncryptedSessionKey: "encrypted_key",
	}

	if msg.MessageID != 1 {
		t.Fatalf("Expected message ID 1, got %d", msg.MessageID)
	}

	if msg.SenderID != 2 {
		t.Fatalf("Expected sender ID 2, got %d", msg.SenderID)
	}

	if msg.SessionID != 3 {
		t.Fatalf("Expected session ID 3, got %d", msg.SessionID)
	}

	if msg.EncryptedSessionKey != "encrypted_key" {
		t.Fatalf("Expected encrypted session key 'encrypted_key', got '%s'", msg.EncryptedSessionKey)
	}
}

// encryptMessageForTest is a helper that encrypts a message for testing decryption
func encryptMessageForTest(message string, sessionKey []byte, recipientPubKey *rsa.PublicKey) (*ChatMessage, error) {
	encryptedSessionKey, err := rsa.EncryptOAEP(sha256New(), rand.Reader, recipientPubKey, sessionKey, nil)
	if err != nil {
		return nil, err
	}

	blockCipher, err := aesNewCipher(sessionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipherNewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(message), nil)

	authTag := ciphertext[len(ciphertext)-16:]
	encryptedContent := ciphertext[:len(ciphertext)-16]

	return &ChatMessage{
		MessageID:           1,
		SenderID:            1,
		SessionID:           1,
		EncryptedSessionKey: base64StdEncodingEncodeToString(encryptedSessionKey),
		EncryptedContent:    base64StdEncodingEncodeToString(encryptedContent),
		Iv:                  base64StdEncodingEncodeToString(nonce),
		AuthTag:             base64StdEncodingEncodeToString(authTag),
	}, nil
}

// Mock error type for testing
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}
