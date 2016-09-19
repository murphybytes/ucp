package main

import (
	"crypto/rsa"
	"os/user"
	"path/filepath"

	"github.com/murphybytes/ucp/crypto"
)

type servicable interface {
	getPrivateKey() *rsa.PrivateKey
	isKeyAuthorized(*user.User, *rsa.PublicKey) (bool, error)
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

func (s *osService) isKeyAuthorized(usr *user.User, publicKey *rsa.PublicKey) (auth bool, e error) {

	return

}
