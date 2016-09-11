package client

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
}

func (s *ClientTestSuite) TestGetEnvironmentFuncs() {
	s.Equal(12345, getIntFromEnvironment("", 12345))
	s.Equal(5555, getIntFromEnvironment("5555", 12345))
}

func TestClientFunctionality(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
