package main

import (
	"io"
	"testing"

	"github.com/murphybytes/ucp/wire"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockFileIntf struct {
	mock.Mock
}

type MockConn struct {
	mock.Mock
}

type MockFile struct {
	mock.Mock
}

func (m *MockFile) Read(b []byte) (n int, e error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockFile) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (fi *MockFileIntf) open(fileName string) (reader io.ReadCloser, e error) {
	args := fi.Called(fileName)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (fi *MockFileIntf) getFileSize() (size int64, e error) {
	args := fi.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (c *MockConn) Read(i interface{}) (e error) {
	args := c.Called(i)
	return args.Error(0)
}

func (c *MockConn) Write(i interface{}) (e error) {
	args := c.Called(i)
	return args.Error(0)
}

type FileSendSuite struct {
	suite.Suite
}

func (s *FileSendSuite) TestFileSend() {
	f := &MockFileIntf{}
	conn := &MockConn{}
	mf := &MockFile{}

	contents := []byte("some file contents")
	fileLen := int64(len(contents))

	txferInfo := wire.FileTransferInformation{
		FileName: "foo",
	}

	f.On("open", txferInfo.FileName).Return(mf, nil)
	mf.On("Close").Return(nil)
	f.On("getFileSize").Return(fileLen, nil)
	txferInfo.FileSize = fileLen
	conn.On("Write", txferInfo).Return(nil)
	mf.On(
		"Read",
		mock.AnythingOfType("[]uint8"),
	).Return(
		int(fileLen),
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).([]byte)
			copy(arg, contents)
		},
	)

	chunk := wire.FileChunk{
		Buffer: contents,
	}

	conn.On("Write", chunk).Return(nil)

	response := wire.FileTransferMore

	conn.On("Read",
		mock.AnythingOfType("*wire.Conversation")).Return(nil).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).(*wire.Conversation)
			*arg = response
		},
	)

	e := fileSend(conn, txferInfo, f)
	s.Nil(e)

}

func TestFileSendSuite(t *testing.T) {
	suite.Run(t, new(FileSendSuite))
}
