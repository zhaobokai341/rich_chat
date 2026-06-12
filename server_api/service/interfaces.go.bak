package service

import (
	"context"
	"errors"
	"time"

	"rich_chat/server_api/database"
)

// Application errors
var (
	ErrUserNotFound             = errors.New("user not found")
	ErrInvalidPassword          = errors.New("invalid password")
	ErrAccountLocked            = errors.New("account is locked")
	ErrInvalidToken             = errors.New("invalid or expired token")
	ErrUsernameAlreadyExists    = errors.New("username already exists")
	ErrInvalidInput             = errors.New("invalid input")
	ErrInvalidEmailFormat       = errors.New("invalid email format")
	ErrEmailExceedsMaxLength    = errors.New("email exceeds maximum length")
	ErrBioExceedsMaxLength      = errors.New("bio exceeds maximum length")
	ErrInvalidNicknameFormat    = errors.New("invalid nickname format")
	ErrPasswordExceedsMaxLength = errors.New("password exceeds maximum length")
)

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string
	Password string
}

// LoginResponse represents successful login response
type LoginResponse struct {
	UserID    int
	UserToken string
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Username string
	Password string
}

// RegisterResponse represents successful registration response
type RegisterResponse struct {
	UserID    int
	UserToken string
}

// UserProfileUpdateRequest represents profile update data
type UserProfileUpdateRequest struct {
	UserID int
	Key    string
	Value  string
}

// DeleteUserRequest represents user deletion request
type DeleteUserRequest struct {
	UserID   int
	Password string
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	UserID      int
	OldPassword string
	NewPassword string
}

// AuthService defines authentication operations
type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	GenerateToken(userID int) (string, error)
	GenerateVerifyToken() (string, error)
	ValidateVerifyToken(token string) error
}

// UserService defines user management operations
type UserService interface {
	GetUserProfile(ctx context.Context, userID int) (*database.UserInfo, error)
	UpdateUserProfile(ctx context.Context, req *UserProfileUpdateRequest) error
	ChangeUserPassword(ctx context.Context, req *ChangePasswordRequest) error
	DeleteUser(ctx context.Context, req *DeleteUserRequest) error
	CheckAccountLocked(identifier string) bool
	CheckUserExists(userID int) (bool, error)
	GetUserBasicInfo(ctx context.Context, userID int) (*database.UserBasicInfo, error)
}

// TokenService defines token operations
type TokenService interface {
	GenerateJWT(userID int, expiration time.Duration) (string, error)
	GenerateVerificationToken() (string, error)
	StoreVerificationToken(token string, ttl time.Duration) error
	ValidateAndConsumeToken(token string) error
}

// CreateSessionRequest represents a request to create a chat session
type CreateSessionRequest struct {
	CreatorID   int
	SessionType string // 'direct', 'group'
	Name        *string
	MemberIDs   []int
}

// CreateSessionResponse represents the response after creating a session
type CreateSessionResponse struct {
	SessionID int
	Name      string
}

// JoinSessionRequest represents a request to join a chat session
type JoinSessionRequest struct {
	UserID    int
	SessionID int
}

// LeaveSessionRequest represents a request to leave a chat session
type LeaveSessionRequest struct {
	UserID    int
	SessionID int
}

// SendMessageRequest represents a request to send an encrypted message
type SendMessageRequest struct {
	SenderID         int
	RecipientID      int
	SessionID        int
	PlaintextContent []byte
	RecipientPubKey  string
}

// SendMessageResponse represents the response after sending a message
type SendMessageResponse struct {
	MessageID           int
	EncryptedSessionKey []byte
	EncryptedContent    []byte
	Iv                  []byte
	AuthTag             []byte
	IsDelivered         bool
}

// GetOfflineMessagesRequest represents a request to get offline messages
type GetOfflineMessagesRequest struct {
	UserID int
	Limit  int // Maximum number of messages to retrieve (0 = no limit)
}

// GetOfflineMessagesResponse represents the response with offline messages
type GetOfflineMessagesResponse struct {
	Messages []*OfflineMessageResponse
	Count    int
}

// OfflineMessageResponse represents a single offline encrypted message
type OfflineMessageResponse struct {
	MessageID           int
	SenderID            int
	SessionID           int
	EncryptedSessionKey []byte
	EncryptedContent    []byte
	Iv                  []byte
	AuthTag             []byte
	CreatedAt           *time.Time
}

// GetUserKeyRequest represents a request to get a user's public key
type GetUserKeyRequest struct {
	UserID int
}

// GetUserKeyResponse represents the response with a user's public key
type GetUserKeyResponse struct {
	UserID    int
	PublicKey string
	Algorithm string
}

// StoreUserKeyRequest represents a request to store a user's public key
// Note: Private key is NEVER stored on server for security reasons
type StoreUserKeyRequest struct {
	UserID       int
	PublicKey    string
	KeyAlgorithm string
}

// SessionInfo represents information about a chat session
type SessionInfo struct {
	SessionID   int        `json:"session_id"`
	SessionType string     `json:"session_type"`
	Name        *string    `json:"name"`
	CreatedAt   *time.Time `json:"created_at"`
	MemberCount int        `json:"member_count"`
	PartnerID   int        `json:"partner_id,omitempty"`   // For direct chats: the other participant's ID
	PartnerName string     `json:"partner_name,omitempty"` // For direct chats: the other participant's name
}

// ChatService defines chat operations with E2EE support
type ChatService interface {
	// Session Management
	CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error)
	JoinSession(ctx context.Context, req *JoinSessionRequest) error
	LeaveSession(ctx context.Context, req *LeaveSessionRequest) error
	GetUserSessions(ctx context.Context, userID int) ([]*SessionInfo, error)

	// E2EE Key Management
	StoreUserKey(ctx context.Context, req *StoreUserKeyRequest) error
	GetUserPublicKey(ctx context.Context, req *GetUserKeyRequest) (*GetUserKeyResponse, error)

	// Message Operations
	SendEncryptedMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error)
	StoreOfflineEncryptedMessage(ctx context.Context, offlineMsg *database.OfflineEncryptedMessage) (int, error)
	GetOfflineMessages(ctx context.Context, req *GetOfflineMessagesRequest) (*GetOfflineMessagesResponse, error)
	MarkMessageDelivered(ctx context.Context, userID, messageID int) error
	GetUndeliveredMessageCount(ctx context.Context, userID int) (int, error)
}
