package pam

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PamTestSuite struct {
	suite.Suite
	userName     string
	userPassword string
}

func (s *PamTestSuite) SetupSuite() {
	// These environment vars must be set in order to test
	// If you are running tests as your
	s.userName = os.Getenv("UCP_TEST_USER")
	s.userPassword = os.Getenv("UCP_TEST_PASSWORD")
	if s.userName == "" || s.userPassword == "" {
		s.T().Fatal("Must set UCP_TEST_USER and UCP_TEST_PASSWORD")
	}
}

func (s *PamTestSuite) TestUserAuthentication() {
	err := AuthorizeUser(s.userName, s.userPassword)
	s.Nil(err)
}

func (s *PamTestSuite) TestUserAuthenticationBadUser() {
	err := AuthorizeUser(s.userName+"xx", s.userPassword)
	s.NotNil(err)
	s.Equal(ErrUnknownUser, err)

}

func (s *PamTestSuite) TestUserAuthenticationBadPasswor() {
	err := AuthorizeUser(s.userName, s.userPassword+"xx")
	s.NotNil(err)
	s.Equal(ErrIncorrectPassword, err)

}

func TestPamTestSuite(t *testing.T) {
	suite.Run(t, new(PamTestSuite))
}
