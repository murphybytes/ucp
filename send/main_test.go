package main

import (
	"testing"

	"github.com/murphybytes/ucp/client"
	"github.com/stretchr/testify/suite"
)

type SenderMainTestSuite struct {
	suite.Suite
}

func (s *SenderMainTestSuite) TestEnvironment() {
	s.NotZero(client.UCPDirectory, "Perhaps .env file is missing. See env.sample")
}

func TestSenderMainTestSuite(t *testing.T) {
	suite.Run(t, new(SenderMainTestSuite))
}
