package server

import "errors"

const DefaultPort = 8978

// subprocess return codes
const Success = 0
const Error = 1

// ErrSocket - Unable to connect to parent process with unix socket
const ErrSocket = 2

const PipeBufferSize = 100000
const FileReaderBufferSize = 10000

var ErrParentTerminatedConversation = errors.New("Connection terminated by parent process")
