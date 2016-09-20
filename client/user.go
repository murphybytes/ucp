package client

import (
	"errors"
	"fmt"

	"github.com/murphybytes/ucp/net"
	"github.com/murphybytes/ucp/wire"
	"golang.org/x/crypto/ssh/terminal"
)

type Prompter interface {
	GetPassword() (string, error)
}

type Prompt struct {
}

func (p *Prompt) GetPassword() (pwd string, e error) {
	fmt.Println("Enter Password: ")
	var buff []byte
	if buff, e = terminal.ReadPassword(0); e != nil {
		return
	}
	pwd = string(buff)
	return
}

func HandleUserAuthorization(conn net.EncodeConn, prompt Prompt) (e error) {
	if e = conn.Write(RemoteUser); e != nil {
		return
	}

	var response wire.UserAuthorizationResponse
	for {
		if e = conn.Read(&response); e != nil {
			return
		}

		switch response.AuthResponse {
		case wire.Authorized:
			return
		case wire.NonexistantUser:
			return errors.New(response.Description)
		case wire.PasswordRequired:
			var pwd string
			if pwd, e = prompt.GetPassword(); e == nil {
				if e = conn.Write(pwd); e != nil {
					return
				}
			} else {
				return
			}
		case wire.IncorrectPassword:
			return errors.New(response.Description)
		}

	}
}