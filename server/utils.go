package server

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
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

func GetUniqueUnixSocketFileName() string {
	buffer := make([]byte, 24)
	rand.Read(buffer)
	name := base32.StdEncoding.EncodeToString(buffer)[:len(buffer)-1]
	return fmt.Sprintf("/tmp/%s.sock", name)

}
