package ui_handler

import "time"

// APIClient defines the interface for API operations
type APIClient interface {
	CheckServerHealth() (bool, error)
	GetVerifyToken() (string, error)
	SetAuthHeaders(token, userID string)
	ClearAuthHeaders()
}

// ChatAPIClient defines the interface for chat API operations
type ChatAPIClient interface {
	GetUserBasicInfo(currentUserID int, token, verifyToken string, targetUserID int) (*UserBasicInfo, error)
	CreateChatSession(userID int, token, verifyToken string, recipientID int) (int, error)
	GetUserSessions(userID int, token, verifyToken string) ([]*SessionInfo, error)
	GetUserPublicKey(currentUserID int, token, verifyToken string, targetUserID int) (string, string, error)
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	IsAuthenticated() bool
	Login(username, password string) error
	Register(username, password string) error
	Logout() error
	GetCredentials() (string, string, error)
}

// UserService defines the interface for user operations
type UserService interface {
	GetProfile() (*UserData, error)
	UpdateProfile(key, value string) error
	ChangePassword(oldPassword, newPassword string) error
	DeleteAccount(password string) error
}

// ChatService defines the interface for chat operations
type ChatService interface {
	Connect(token string, userID int) error
	Disconnect() error
	HasEncryptionKey(userID int) (bool, error)
	SaveEncryptionKey(userID int, key string) error
	LoadEncryptionKey(userID int) error
	GenerateAndUploadKeys(userID int, token, verifyToken string) error
	UploadPublicKeyOnly(userID int, token, verifyToken string) error
	SendMessage(sessionID, recipientID int, message, recipientPubKeyPEM string) error
	GetMessages() <-chan *ChatMessage
	DecryptMessage(msg *ChatMessage) (string, error)
}

// ConfigManager defines the interface for configuration operations
type ConfigManager interface {
	ReadConfig() (map[string]interface{}, error)
	GetToken() (string, bool)
	GetUserID() (string, bool)
}

// LanguagePackWrapper defines the interface for language operations
type LanguagePack interface {
	Get(key string) string
}

// UserBasicInfo represents basic user information (from chat API)
type UserBasicInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
}

// UserData represents user profile data (from user API)
type UserData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
}

// SessionInfo represents chat session information
type SessionInfo struct {
	SessionID   int
	PartnerID   int
	PartnerName string
}

// ChatMessage represents an encrypted chat message
type ChatMessage struct {
	MessageID           int       `json:"message_id"`
	SenderID            int       `json:"sender_id"`
	SessionID           int       `json:"session_id"`
	EncryptedSessionKey string    `json:"encrypted_session_key"`
	EncryptedContent    string    `json:"encrypted_content"`
	Iv                  string    `json:"iv"`
	AuthTag             string    `json:"auth_tag"`
	CreatedAt           time.Time `json:"created_at"`
}

// Config represents application configuration
type Config struct {
	Dir string
}

// Constants
const (
	DefaultConfigDir = "config"
)
