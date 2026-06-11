package main

import (
	"rich_chat/client/ui_handler"
)

// userServiceAdapter adapts UserService to ui_handler.UserService interface
type userServiceAdapter struct {
	service *UserService
}

func (a *userServiceAdapter) GetProfile() (*ui_handler.UserData, error) {
	data, err := a.service.GetProfile()
	if err != nil {
		return nil, err
	}
	return &ui_handler.UserData{
		Username: data.Username,
		Email:    data.Email,
		Nickname: data.Nickname,
		Bio:      data.Bio,
	}, nil
}

func (a *userServiceAdapter) UpdateProfile(key, value string) error {
	return a.service.UpdateProfile(key, value)
}

func (a *userServiceAdapter) ChangePassword(oldPassword, newPassword string) error {
	return a.service.ChangePassword(oldPassword, newPassword)
}

func (a *userServiceAdapter) DeleteAccount(password string) error {
	return a.service.DeleteAccount(password)
}

// chatAPIClientAdapter adapts ChatAPIClient to ui_handler.ChatAPIClient interface
type chatAPIClientAdapter struct {
	client *ChatAPIClient
}

func (a *chatAPIClientAdapter) GetUserBasicInfo(currentUserID int, token, verifyToken string, targetUserID int) (*ui_handler.UserBasicInfo, error) {
	info, err := a.client.GetUserBasicInfo(currentUserID, token, verifyToken, targetUserID)
	if err != nil {
		return nil, err
	}
	return &ui_handler.UserBasicInfo{
		ID:       info.ID,
		Username: info.Username,
		Nickname: info.Nickname,
		Bio:      info.Bio,
	}, nil
}

func (a *chatAPIClientAdapter) CreateChatSession(userID int, token, verifyToken string, recipientID int) (int, error) {
	return a.client.CreateChatSession(userID, token, verifyToken, recipientID)
}

func (a *chatAPIClientAdapter) GetUserSessions(userID int, token, verifyToken string) ([]*ui_handler.SessionInfo, error) {
	sessions, err := a.client.GetUserSessions(userID, token, verifyToken)
	if err != nil {
		return nil, err
	}
	result := make([]*ui_handler.SessionInfo, len(sessions))
	for i, s := range sessions {
		result[i] = &ui_handler.SessionInfo{
			SessionID:   s.SessionID,
			PartnerID:   s.PartnerID,
			PartnerName: s.PartnerName,
		}
	}
	return result, nil
}

func (a *chatAPIClientAdapter) GetUserPublicKey(currentUserID int, token, verifyToken string, targetUserID int) (string, string, error) {
	return a.client.GetUserPublicKey(currentUserID, token, verifyToken, targetUserID)
}

// chatServiceAdapter adapts ChatService to ui_handler.ChatService interface
type chatServiceAdapter struct {
	service *ChatService
}

func (a *chatServiceAdapter) Connect(token string, userID int) error {
	return a.service.Connect(token, userID)
}

func (a *chatServiceAdapter) Disconnect() error {
	return a.service.Disconnect()
}

func (a *chatServiceAdapter) HasEncryptionKey(userID int) (bool, error) {
	return a.service.HasEncryptionKey(userID)
}

func (a *chatServiceAdapter) SaveEncryptionKey(userID int, key string) error {
	return a.service.SaveEncryptionKey(userID, key)
}

func (a *chatServiceAdapter) LoadEncryptionKey(userID int) error {
	return a.service.LoadEncryptionKey(userID)
}

func (a *chatServiceAdapter) GenerateAndUploadKeys(userID int, token, verifyToken string) error {
	return a.service.GenerateAndUploadKeys(userID, token, verifyToken)
}

func (a *chatServiceAdapter) UploadPublicKeyOnly(userID int, token, verifyToken string) error {
	return a.service.UploadPublicKeyOnly(userID, token, verifyToken)
}

func (a *chatServiceAdapter) SendMessage(sessionID, recipientID int, message, recipientPubKeyPEM string) error {
	return a.service.SendMessage(sessionID, recipientID, message, recipientPubKeyPEM)
}

func (a *chatServiceAdapter) GetMessages() <-chan *ui_handler.ChatMessage {
	out := make(chan *ui_handler.ChatMessage)
	in := a.service.GetMessages()
	go func() {
		for msg := range in {
			out <- &ui_handler.ChatMessage{
				MessageID:           msg.MessageID,
				SenderID:            msg.SenderID,
				SessionID:           msg.SessionID,
				EncryptedSessionKey: msg.EncryptedSessionKey,
				EncryptedContent:    msg.EncryptedContent,
				Iv:                  msg.Iv,
				AuthTag:             msg.AuthTag,
				CreatedAt:           msg.CreatedAt,
			}
		}
		close(out)
	}()
	return out
}

func (a *chatServiceAdapter) DecryptMessage(msg *ui_handler.ChatMessage) (string, error) {
	// Convert ui_handler.ChatMessage to client.ChatMessage
	clientMsg := &ChatMessage{
		MessageID:           msg.MessageID,
		SenderID:            msg.SenderID,
		SessionID:           msg.SessionID,
		EncryptedSessionKey: msg.EncryptedSessionKey,
		EncryptedContent:    msg.EncryptedContent,
		Iv:                  msg.Iv,
		AuthTag:             msg.AuthTag,
		CreatedAt:           msg.CreatedAt,
	}
	return a.service.DecryptMessage(clientMsg)
}
