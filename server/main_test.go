package main

import (
	"crypto/rand"
	"crypto/rsa"
	"os/user"
	"testing"

	"github.com/murphybytes/ucp/crypto"
	"github.com/murphybytes/ucp/wire"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockEncodeConn struct {
	mock.Mock
}

type MockServiceable struct {
	mock.Mock
}

func (ms *MockServiceable) getPrivateKey() *rsa.PrivateKey {
	args := ms.Called()
	return args.Get(0).(*rsa.PrivateKey)
}

func (ms *MockServiceable) isKeyAuthorized(u *user.User, key []byte, authFile func() []byte) (bool, error) {
	args := ms.Called(u, key, authFile)
	return args.Bool(0), args.Error(1)
}

func (ms *MockServiceable) lookupUser(userName string) (*user.User, error) {
	args := ms.Called(userName)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockEncodeConn) Read(a interface{}) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *MockEncodeConn) Write(a interface{}) error {
	args := m.Called(a)
	return args.Error(0)
}

//func(n *NetworkMock) Read()

type ServerMainTestSuite struct {
	suite.Suite
	conn            *MockEncodeConn
	service         *MockServiceable
	clientPublicKey *rsa.PublicKey
}

func (s *ServerMainTestSuite) SetupTest() {
	s.conn = new(MockEncodeConn)
	s.service = new(MockServiceable)
	privateKey, _ := rsa.GenerateKey(rand.Reader, crypto.KeySize)
	s.clientPublicKey = &privateKey.PublicKey

}

func (s *ServerMainTestSuite) TestCreateHandleUserAuthorization() {
	userName := "bob"
	expectedUser := user.User{
		Username: userName,
	}

	s.conn.On("Write", wire.UserNameRequest).Return(nil)
	s.conn.On(
		"Read",
		mock.AnythingOfType("*string"),
	).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*string)
		*arg = userName
	})

	s.service.On(
		"lookupUser",
		userName,
	).Return(
		&expectedUser,
		nil,
	)

	s.service.On(
		"isKeyAuthorized",
		&expectedUser,
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("func() []uint8"),
	).Return(
		true,
		nil,
	)

	s.conn.On(
		"Write",
		wire.UserAuthorizationResponse{
			AuthResponse: wire.Authorized,
		},
	).Return(
		nil,
	)

	user, err := handleUserAuthorization(s.conn, s.service, s.clientPublicKey)
	s.Nil(err)
	s.NotNil(user)

}

func TestServerMainTestSuite(t *testing.T) {
	suite.Run(t, new(ServerMainTestSuite))
}
