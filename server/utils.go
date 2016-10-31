package server

import (
	"bytes"
	"fmt"
	"io"
)

const readBufferSize = PipeBufferSize * 2

type ReadWriteJoiner struct {
	reader io.Reader
	writer io.Writer
}

func NewReadWriteJoiner(r io.Reader, w io.Writer) *ReadWriteJoiner {
	return &ReadWriteJoiner{
		reader: r,
		writer: w,
	}
}

func (s *ReadWriteJoiner) Read(b *bytes.Buffer) (e error) {

	buffer := make([]byte, readBufferSize)
	for {

		var read int
		read, e = s.reader.Read(buffer)
		fmt.Printf("Read %d Error %+v\n", read, e)
		if e != nil && e != io.EOF {
			return
		}

		if read > 0 {
			b.Write(buffer[:read])
			//		continue
		}

		if read < readBufferSize {
			fmt.Println("Exit")
			return
		}

	}
}

func (s *ReadWriteJoiner) Write(b []byte) (n int, e error) {
	return s.writer.Write(b)
}

func (s *ReadWriteJoiner) Close() (e error) {
	return
}
