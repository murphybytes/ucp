package server

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockReader struct {
	mock.Mock
}

func (r *MockReader) Read(b []byte) (n int, e error) {
	args := r.Called(b)
	return args.Int(0), args.Error(1)
}

type MockWriter struct {
	mock.Mock
}

func (w *MockWriter) Write(b []byte) (n int, e error) {
	args := w.Called(b)
	return args.Int(0), args.Error(1)
}

type UtilsSuite struct {
	suite.Suite
	reader *MockReader
	writer *MockWriter
}

func (s *UtilsSuite) SetupTest() {
	s.reader = &MockReader{}
}

func (s *UtilsSuite) TestReadWriteJoiner() {
	undertest := NewReadWriteJoiner(s.reader, s.writer)

	buff := make([]byte, PipeBufferSize*1.5)
	rand.Read(buff)

	s.reader.On(
		"Read",
		mock.AnythingOfType("[]uint8"),
	).Return(
		len(buff[:PipeBufferSize]),
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).([]byte)
			copy(arg, buff)
		},
	).Once()

	s.reader.On(
		"Read",
		mock.AnythingOfType("[]uint8"),
	).Return(
		len(buff[PipeBufferSize:]),
		nil,
	).Run(
		func(args mock.Arguments) {
			arg := args.Get(0).([]byte)
			copy(arg, buff[PipeBufferSize:])
		},
	)

	var reader bytes.Buffer
	err := undertest.Read(&reader)
	s.Nil(err)
	s.Equal(len(buff), reader.Len())
	s.True(bytes.Equal(reader.Bytes(), buff))

}

func TestUtilsSuite(t *testing.T) {
	suite.Run(t, new(UtilsSuite))
}
