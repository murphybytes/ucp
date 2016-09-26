package main

import (
	"bufio"
	"bytes"
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/bgentry/speakeasy"
	"github.com/murphybytes/ucp/crypto"
)

type servicable interface {
	getPrivateKey() *rsa.PrivateKey
	isKeyAuthorized(*user.User, []byte, func() []byte) (bool, error)
	lookupUser(string) (*user.User, error)
	validatePassword(*user.User, string) error
}

type userLookupFunc func(string) (*user.User, error)

// osServices wraps os functionality, file access act
type osService struct {
	privateKey *rsa.PrivateKey
}

func newOsService() (service *osService, e error) {
	service = &osService{}
	service.privateKey, e = crypto.GetPrivateKey(filepath.Join(ucpDirectory, "private-key.pem"))
	if e != nil {
		return
	}

	return
}

func (s *osService) getPrivateKey() (key *rsa.PrivateKey) {
	return s.privateKey
}

func (s *osService) isKeyAuthorized(usr *user.User, encodedKey []byte,
	authfile func() []byte) (auth bool, e error) {

	contents := authfile()

	// strip off line feed
	encodedKey = encodedKey[:len(encodedKey)-1]

	for _, line := range bytes.Split(contents, []byte{'\n'}) {

		if bytes.Equal(line, encodedKey) {
			return true, nil
		}
	}

	return false, nil

}

func (s *osService) lookupUser(userName string) (u *user.User, e error) {
	return user.Lookup(userName)
}

type PamHandler struct {
	Password string
}

func (p *PamHandler) RespondPAM(msgStyle int, msg string) (string, bool) {
	return p.Password, true
}

func (s osService) validatePassword(user *user.User, password string) error {

	t, err := pam.StartFunc("", user.Username, func(s pam.Style, msg string) (string, error) {
		fmt.Println("callback")
		switch s {
		case pam.PromptEchoOff:
			fmt.Println("echo off")
			return speakeasy.Ask(msg)
		case pam.PromptEchoOn:
			fmt.Println("echo on")
			fmt.Print(msg + " ")
			input, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return "", err
			}
			return input[:len(input)-1], nil
		case pam.ErrorMsg:
			log.Print(msg)
			return "", nil
		case pam.TextInfo:
			fmt.Println(msg)
			return "", nil
		}
		return "", errors.New("Unrecognized message style")
	})

	if err != nil {
		return err
	}

	err = t.Authenticate(pam.Silent)

	return err

}
