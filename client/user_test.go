package client

import (
	"testing"

	"github.com/murphybytes/ucp/wire"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockConnection struct {
	mock.Mock
}

func (mc *MockConnection) Write(v interface{}) error {
	args := mc.Called(v)
	return args.Error(0)
}

func (mc *MockConnection) Read(v interface{}) error {
	args := mc.Called(v)
	return args.Error(0)
}

type MockPrompt struct {
	mock.Mock
}

func (mp *MockPrompt) GetPassword() (pwd string, e error) {
	args := mp.Called()
	return args.String(0), args.Error(1)
}

type UserTestSuite struct {
	suite.Suite
	prompt *MockPrompt
	conn   *MockConnection
}

func (s *UserTestSuite) SetupTest() {
	s.prompt = &MockPrompt{}
	s.conn = &MockConnection{}

}

func (s *UserTestSuite) TestAuthWithAuthorizedUser() {

	s.conn.On(
		"Read",
		mock.AnythingOfType("*wire.Conversation"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.Conversation)
			*arg = wire.UserNameRequest
		},
	)

	s.conn.On(
		"Write",
		mock.AnythingOfType("string"),
	).Return(
		nil,
	)

	s.conn.On(
		"Read",
		mock.AnythingOfType("*wire.UserAuthorizationResponse"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.UserAuthorizationResponse)
			*arg = wire.UserAuthorizationResponse{
				AuthResponse: wire.Authorized,
			}
		},
	)

	e := HandleUserAuthorization(s.conn, s.prompt)
	s.Nil(e)

}

func (s *UserTestSuite) TestAuthWithNonAuthorizedUser() {

	s.conn.On(
		"Read",
		mock.AnythingOfType("*wire.Conversation"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.Conversation)
			*arg = wire.UserNameRequest
		},
	)

	s.conn.On(
		"Write",
		mock.AnythingOfType("string"),
	).Return(
		nil,
	).Once()

	s.conn.On(
		"Read",
		mock.AnythingOfType("*wire.UserAuthorizationResponse"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.UserAuthorizationResponse)
			*arg = wire.UserAuthorizationResponse{
				AuthResponse: wire.PasswordRequired,
			}
		},
	).Once()

	password := "somePassword"

	s.prompt.On(
		"GetPassword",
	).Return(
		password,
		nil,
	)

	s.conn.On(
		"Write",
		password,
	).Return(
		nil,
	)

	s.conn.On(
		"Read",
		mock.AnythingOfType("*wire.UserAuthorizationResponse"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.UserAuthorizationResponse)
			*arg = wire.UserAuthorizationResponse{
				AuthResponse: wire.Authorized,
			}
		},
	)

	e := HandleUserAuthorization(s.conn, s.prompt)
	s.Nil(e)

}

func (s *UserTestSuite) TestAuthWithIncorrectPassword() {

	s.conn.On(
		"Read",
		mock.AnythingOfType("*wire.Conversation"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.Conversation)
			*arg = wire.UserNameRequest
		},
	)

	s.conn.On(
		"Write",
		mock.AnythingOfType("string"),
	).Return(
		nil,
	).Once()

	s.conn.On(
		"Read",
		mock.AnythingOfType("*wire.UserAuthorizationResponse"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.UserAuthorizationResponse)
			*arg = wire.UserAuthorizationResponse{
				AuthResponse: wire.PasswordRequired,
			}
		},
	).Once()

	password := "somePassword"

	s.prompt.On(
		"GetPassword",
	).Return(
		password,
		nil,
	)

	s.conn.On(
		"Write",
		password,
	).Return(
		nil,
	)

	description := "Incorrect Password"

	s.conn.On(
		"Read",
		mock.AnythingOfType("*wire.UserAuthorizationResponse"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.UserAuthorizationResponse)
			*arg = wire.UserAuthorizationResponse{
				AuthResponse: wire.IncorrectPassword,
				Description:  description,
			}
		},
	)

	e := HandleUserAuthorization(s.conn, s.prompt)
	s.NotNil(e)
	s.Equal(description, e.Error())

}

func (s *UserTestSuite) TestAuthWithNonExistantUser() {

	s.conn.On(
		"Read",
		mock.AnythingOfType("*wire.Conversation"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.Conversation)
			*arg = wire.UserNameRequest
		},
	)

	s.conn.On(
		"Write",
		mock.AnythingOfType("string"),
	).Return(
		nil,
	).Once()

	s.conn.On(
		"Read",
		mock.AnythingOfType("*wire.UserAuthorizationResponse"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.UserAuthorizationResponse)
			*arg = wire.UserAuthorizationResponse{
				AuthResponse: wire.PasswordRequired,
			}
		},
	).Once()

	password := "somePassword"

	s.prompt.On(
		"GetPassword",
	).Return(
		password,
		nil,
	)

	s.conn.On(
		"Write",
		password,
	).Return(
		nil,
	)

	description := "No user"

	s.conn.On(
		"Read",
		mock.AnythingOfType("*wire.UserAuthorizationResponse"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.UserAuthorizationResponse)
			*arg = wire.UserAuthorizationResponse{
				AuthResponse: wire.NonexistantUser,
				Description:  description,
			}
		},
	)

	e := HandleUserAuthorization(s.conn, s.prompt)
	s.NotNil(e)
	s.Equal(description, e.Error())

}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
