package main

import (
	"bytes"
	"crypto/rsa"
	"os/user"
	"path/filepath"

	"github.com/murphybytes/ucp/crypto"
)

type servicable interface {
	getPrivateKey() *rsa.PrivateKey
	isKeyAuthorized(*user.User, []byte, func() []byte) (bool, error)
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
