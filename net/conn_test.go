package net

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/murphybytes/ucp/crypto"
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

	m *mockConn
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
	copy(readBuffer, send)

	s.m.On(
		"Write",
		send,
	).Return(
		len(send),
		nil,
	)

	s.writer.Write(buffer)
	e := s.writer.Read(&reader)
	s.Nil(e)
	s.Equal(len(buffer), reader.Len())
	s.Equal(0, bytes.Compare(buffer, reader.Bytes()))

}

func TestRunConnTestSuite(t *testing.T) {
	suite.Run(t, new(ConnTestSuite))
}

type mockReaderWriter struct {
	mock.Mock
	buffer []byte
}

func (m *mockReaderWriter) Write(buffer []byte) (n int, e error) {
	m.buffer = buffer
	return len(buffer), nil
}

func (m *mockReaderWriter) Read(buff *bytes.Buffer) (e error) {
	buff.Write(m.buffer)
	return nil
}

func (m *mockReaderWriter) Close() (e error) {
	args := m.Called()
	return args.Error(0)
}

type EncryptorTestSuite struct {
	suite.Suite
	m                     *mockReaderWriter
	encryptedReaderWriter Conn
	block                 cipher.Block
	iv                    []byte
}

func (s *EncryptorTestSuite) SetupTest() {
	s.m = new(mockReaderWriter)
	s.iv = make([]byte, crypto.IVBlockSize)
	rand.Read(s.iv)
	s.block, _ = crypto.NewCipherBlock()
	s.encryptedReaderWriter = NewCryptoReaderWriter(s.block, s.iv, s.m)

}

func (s *EncryptorTestSuite) TestEncryptedReadAndWrite() {
	inbuff := []byte("I am some text")

	var outbuff bytes.Buffer

	n, e := s.encryptedReaderWriter.Write(inbuff)
	s.Nil(e)
	s.Equal(len(inbuff), n)

	expected := crypto.EncryptAES(s.block, s.iv, inbuff)
	s.Equal(string(expected), string(s.m.buffer))

	e = s.encryptedReaderWriter.Read(&outbuff)
	s.Nil(e)
	s.Equal("I am some text", string(outbuff.Bytes()))

}

func (s *EncryptorTestSuite) TestGobEncodedReadAndWrite() {
	expected := struct {
		Question string
		Answer   int
	}{
		Question: "What is the Answer to everything?",
		Answer:   42,
	}

	rw := NewGobEncoderReaderWriter(s.encryptedReaderWriter)

	e := rw.Write(expected)
	s.Nil(e)

	actual := struct {
		Question string
		Answer   int
	}{}

	e = rw.Read(&actual)
	s.Nil(e)
	s.Equal(expected.Answer, actual.Answer)
	s.Equal(expected.Question, actual.Question)

}

func (s *EncryptorTestSuite) TestGobEncodedReadAndWriteWithString() {
	expected := "Here is a string"

	rw := NewGobEncoderReaderWriter(s.encryptedReaderWriter)

	e := rw.Write(expected)
	s.Nil(e)

	actual := ""

	e = rw.Read(&actual)
	s.Nil(e)
	s.Equal(expected, actual)

}

func TestRunEncryptorTestSuite(t *testing.T) {
	suite.Run(t, new(EncryptorTestSuite))
}

type RSAEncryptionTestSuite struct {
	suite.Suite
	rsaReaderWriter Conn
	m               *mockReaderWriter
}

func (s *RSAEncryptionTestSuite) SetupTest() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, crypto.KeySize)

	s.m = new(mockReaderWriter)

	s.rsaReaderWriter = NewRSAReaderWriter(&privateKey.PublicKey, privateKey, s.m)
}

func (s *RSAEncryptionTestSuite) TestRSAEncryptedReadWrite() {
	inbuff := []byte("Here are some words")
	_, e := s.rsaReaderWriter.Write(inbuff)
	s.Nil(e)
	//	s.Equal(len(inbuff), n)

	var outbuff bytes.Buffer
	e = s.rsaReaderWriter.Read(&outbuff)
	s.Nil(e)
	s.True(bytes.Equal(inbuff, outbuff.Bytes()))
}

func TestRSAEncryptionTestSuiteTest(t *testing.T) {
	suite.Run(t, new(RSAEncryptionTestSuite))
}
