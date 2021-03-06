package wire

import "errors"

var ErrUnauthorizedUser = errors.New("Unauthorized user")

type Conversation string

const (
	UserNameRequest                Conversation = "USER_NAME"
	FileTransferInformationRequest Conversation = "REQUEST_FILE_TRANSFER_INFORMATION"
	FileTransferStart              Conversation = "FILE_TRANSFER_START"
	FileTransferSuccess            Conversation = "FILE_TRANSFER_SUCCESS"
	FileTransferFail               Conversation = "FILE_TRANSFER_FAIL"
	FileTransferAbort              Conversation = "FILE_TRANSFER_ABORT"
	FileTransferMore               Conversation = "FILE_TRANSFER_MORE"
	FileTransferComplete           Conversation = "FILE_TRANSFER_COMPLETE"
)

// SymmetricEncryptionParms contains values used for AES encryption
// that are generated by server
type SymmetricEncryptionParms struct {
	Key                  []byte
	InitializationVector []byte
	ClientAck            bool
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

type TransferType int

const (
	FileSend TransferType = iota
	FileReceive
)

type FileTransferInformation struct {
	FileTransferType TransferType
	FileName         string
	FileSize         int64
	Error            error
}

type FileChunk struct {
	Buffer []byte
	Error  error
}
