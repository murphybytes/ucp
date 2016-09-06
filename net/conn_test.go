package net

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type mockConn struct {
	mock.Mock
	buffer []byte
}

func (m *mockConn) Write(buffer []byte) (n int, e error) {
	m.buffer = buffer
	args := m.Called(buffer)
	return args.Int(0), args.Error(1)
}

func (m *mockConn) Read(buffer []byte) (n int, e error) {
	n = copy(buffer, m.buffer)
	if len(buffer) <= len(m.buffer) {
		m.buffer = m.buffer[n:]
	} else {
		m.buffer = []byte{}
	}

	return
}

func (m *mockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

type ConnTestSuite struct {
	suite.Suite
	writer Conn
	m      *mockConn
}

func (s *ConnTestSuite) SetupTest() {
	s.m = new(mockConn)
	s.writer = NewReaderWriter(s.m)

}

func createPacket(buff []byte) (p []byte) {
	p, _ = buildPacket(buff)
	return
}

func (s *ConnTestSuite) TestWriter() {
	buffer := make([]byte, 100)
	rand.Read(buffer)
	toSend := createPacket(buffer)

	s.m.On(
		"Write",
		toSend,
	).Return(
		len(toSend),
		nil,
	)

	n, e := s.writer.Write(buffer)
	s.Nil(e)
	s.Equal(len(buffer), n)
	s.m.AssertExpectations(s.T())

}

func TestRunConnTestSuite(t *testing.T) {
	suite.Run(t, new(ConnTestSuite))
}
