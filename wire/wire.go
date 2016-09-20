package wire

import (
	"crypto/cipher"
)

type Conversation string

const (
	UserNameRequest Conversation = "USER_NAME"
)

// SymmetricEncryptionParms contains values used for AES encryption
// that are generated by server
type SymmetricEncryptionParms struct {
	Block                cipher.Block
	InitializationVector []byte
}

type AuthorizationCode int

const (
	Authorized AuthorizationCode = iota
	PasswordRequired
	NonexistantUser
	IncorrectPassword
)

// UserAuthorizationResponse indicated if user is authorized or not
type UserAuthorizationResponse struct {
	AuthResponse AuthorizationCode
	Description  string
}