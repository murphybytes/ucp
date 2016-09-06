package net

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
)

const readSize = 1500

var ErrInvalidChecksum = errors.New("Invalid checksum")

type Conn interface {
	Write(b []byte) (int, error)
	Close() error
	Read(*bytes.Buffer) (int, error)
}

type ReaderWriter struct {
	conn io.ReadWriteCloser
}

func NewReaderWriter(conn io.ReadWriteCloser) (w *ReaderWriter) {
	return &ReaderWriter{
		conn: conn,
	}
}

func (w *ReaderWriter) Write(buffer []byte) (n int, e error) {
	var packet []byte
	if packet, e = buildPacket(buffer); e != nil {
		return
	}

	if n, e = w.conn.Write(packet); e != nil {
		return
	}

	n -= headerSize

	if n != len(buffer) {
		e = fmt.Errorf("Incomplete write. Wrote %d, expected %d\n", n, len(buffer))
	}

	return
}

func (w *ReaderWriter) Close() error {
	return w.conn.Close()
}

func (w *ReaderWriter) Read(out *bytes.Buffer) (n int, e error) {
	buff := make([]byte, readSize)
	var r int
	r, e = w.conn.Read(buff)

	if e != nil && e != io.EOF {
		return
	}

	if r == 0 {
		return r, io.EOF
	}

	var size int
	if size, e = getBufferSize(buff); e != nil {
		return
	}

	expectedChecksum := buff[sizeHeaderSize : sizeHeaderSize+md5.Size]

	out.Write(buff[headerSize : r-headerSize])
	n = out.Len()

	for n < size {
		r, e = w.conn.Read(buff)
		if e != nil && e != io.EOF {
			return
		}

		out.Write(buff[0:r])
		n += r

		if e == io.EOF {
			break
		}

	}

	actualChecksum := md5.Sum(out.Bytes())
	if bytes.Compare(expectedChecksum[:], actualChecksum[:]) != 0 {
		e = ErrInvalidChecksum
	}

	return out.Len(), e
}
