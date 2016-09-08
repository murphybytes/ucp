package net

import (
	"bytes"
	"crypto/cipher"
	"crypto/md5"
	"errors"
	"io"

	"github.com/murphybytes/ucp/crypto"
)

const readSize = 1500

var ErrInvalidChecksum = errors.New("Invalid checksum")
var ErrIncompleteWrite = errors.New("Incomplete  write")

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
	if packet, e = prependHeaderToBuffer(buffer); e != nil {
		return
	}

	if n, e = w.conn.Write(packet); e != nil {
		return
	}

	if n != len(packet) {
		e = ErrIncompleteWrite
	}

	n -= headerLen

	return

}

func (w *ReaderWriter) Close() error {
	return w.conn.Close()
}

func (w *ReaderWriter) Read(out *bytes.Buffer) (n int, e error) {
	buff := make([]byte, readSize)
	var read int
	read, e = w.conn.Read(buff)

	if e != nil && e != io.EOF {
		return
	}

	if read == 0 {
		return read, io.EOF
	}

	var size int
	if size, e = getBufferSize(buff); e != nil {
		return
	}

	var expectedChecksum [checksumHeaderLen]byte
	copy(expectedChecksum[:], buff[sizeHeaderLen:sizeHeaderLen+checksumHeaderLen])

	out.Write(buff[headerLen:read])
	n = out.Len()

	for n < size {
		read, e = w.conn.Read(buff)
		if e != nil && e != io.EOF {
			return
		}

		out.Write(buff[0:read])
		n += read

		if e == io.EOF {
			break
		}

	}

	actualChecksum := md5.Sum(out.Bytes())

	if !bytes.Equal(expectedChecksum[0:], actualChecksum[0:]) {
		e = ErrInvalidChecksum
	}

	return out.Len(), e
}

type CryptoReaderWriter struct {
	readerWriter         *ReaderWriter
	block                cipher.Block
	initializationVector []byte
}

func NewCryptoReaderWriter(block cipher.Block, initializationVector []byte, readerWriter *ReaderWriter) (rw *CryptoReaderWriter) {
	return &CryptoReaderWriter{
		readerWriter:         readerWriter,
		initializationVector: initializationVector,
	}
}

func (crw *CryptoReaderWriter) Close() error {
	return crw.readerWriter.Close()
}

func (crw *CryptoReaderWriter) Read(buff *bytes.Buffer) (n int, e error) {
	var encrypted bytes.Buffer
	if n, e = crw.readerWriter.Read(&encrypted); e != nil {
		return
	}

	decrypted := crypto.DecryptAES(crw.block, crw.initializationVector, encrypted.Bytes())
	buff.Write(decrypted)
	return
}

func (crw *CryptoReaderWriter) Write(buff []byte) (n int, e error) {
	encrypted := crypto.EncryptAES(crw.block, crw.initializationVector, buff)
	return crw.readerWriter.Write(encrypted)
}
