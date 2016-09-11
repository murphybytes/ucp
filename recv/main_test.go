package main

import (
	"testing"

	"github.com/murphybytes/ucp/client"
	"github.com/stretchr/testify/suite"
)

type RecvMainTestSuite struct {
	suite.Suite
}

func (s *RecvMainTestSuite) TestEnvironment() {
	s.NotZero(client.UCPDirectory, "Perhaps .env file is missing. See env.sample")
}

func TestRecvMainTestSuite(t *testing.T) {
	suite.Run(t, new(RecvMainTestSuite))
}
