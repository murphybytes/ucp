package net

import (
	"bytes"
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

	if len(m.buffer) > len(buffer) {
		m.buffer = m.buffer[n:]
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
	p, _ = prependHeaderToBuffer(buff)
	return
}

func (s *ConnTestSuite) TestWriter() {
	buffer := make([]byte, 50)
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

func (s *ConnTestSuite) TestReader() {
	buffer := make([]byte, 3600)
	readBuffer := make([]byte, readSize)

	var reader bytes.Buffer

	rand.Read(buffer)
	send := createPacket(buffer)
	s.m.On(
		"Write",
		send,
	).Return(
		len(send),
		nil,
	)

	s.m.On(
		"Read",
		readBuffer,
	).Return(
		len(buffer),
		nil,
	)

	s.writer.Write(buffer)
	n, e := s.writer.Read(&reader)
	s.Nil(e)
	s.Equal(len(buffer), reader.Len())
	s.Equal(len(buffer), n)
	s.Equal(0, bytes.Compare(buffer, reader.Bytes()))

}

func TestRunConnTestSuite(t *testing.T) {
	suite.Run(t, new(ConnTestSuite))
}
