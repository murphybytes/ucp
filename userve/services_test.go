package main

import (
	"crypto/rand"
	"crypto/rsa"
	"os/user"
	"testing"

	uc "github.com/murphybytes/ucp/crypto"
	"github.com/stretchr/testify/suite"
)

type ServicesTestSuite struct {
	suite.Suite
	service        *osService
	encodedKey     []byte
	authorizedKeys []byte
}

func (s *ServicesTestSuite) SetupTest() {
	s.service = &osService{}
	s.service.privateKey, _ = rsa.GenerateKey(rand.Reader, uc.KeySize)
	s.authorizedKeys, _ = uc.CreateBase64EncodedPublicKey(s.service.privateKey)

	p, _ := rsa.GenerateKey(rand.Reader, uc.KeySize)
	b, _ := uc.CreateBase64EncodedPublicKey(p)
	s.authorizedKeys = append(s.authorizedKeys, b...)

	s.encodedKey, _ = uc.CreateBase64EncodedPublicKey(s.service.privateKey)

}

func (s *ServicesTestSuite) TestAuthorizedKeyPresent() {
	// dead fox
	var u user.User
	auth, e := s.service.isKeyAuthorized(&u, s.encodedKey, func() []byte { return s.authorizedKeys })
	s.Nil(e)
	s.True(auth)

}

func (s *ServicesTestSuite) TestAuthorizedKeyNotPresent() {

	someKeys := []byte{}

	for i := 0; i < 3; i++ {
		p, _ := rsa.GenerateKey(rand.Reader, uc.KeySize)
		b, _ := uc.CreateBase64EncodedPublicKey(p)
		someKeys = append(someKeys, b...)
	}

	var u user.User
	auth, e := s.service.isKeyAuthorized(&u, s.encodedKey, func() []byte { return someKeys })
	s.Nil(e)
	s.False(auth)

}

func (s *ServicesTestSuite) TestAuthorizedKeyFileEmpty() {

	var u user.User
	auth, e := s.service.isKeyAuthorized(&u, s.encodedKey, func() []byte { return []byte{} })
	s.Nil(e)
	s.False(auth)

}

func TestServicesTestSuite(t *testing.T) {
	suite.Run(t, new(ServicesTestSuite))
}
