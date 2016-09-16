package net

import (
	"bytes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rsa"
	"encoding/gob"
	"errors"
	"fmt"
	"io"

	"github.com/murphybytes/ucp/crypto"
)

const readSize = 1500

var ErrInvalidChecksum = errors.New("Invalid checksum")
var ErrIncompleteWrite = errors.New("Incomplete  write")

// Conn io interface for ucp
type Conn interface {
	Write(b []byte) (int, error)
	Read(*bytes.Buffer) error
}

// ReaderWriter reads and writes packets of bytes to network
// bytes are prepended with size and checksum
type ReaderWriter struct {
	conn io.ReadWriteCloser
}

func NewReaderWriter(conn io.ReadWriteCloser) (w *ReaderWriter) {
	return &ReaderWriter{
		conn: conn,
	}
}

// Write writes buffer to network. Returns the number of bytes written
// if successful, otherwise an error.
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

// Read reads bytes into bytes.Buffer
func (w *ReaderWriter) Read(out *bytes.Buffer) (e error) {
	buff := make([]byte, readSize)
	var read int
	read, e = w.conn.Read(buff)

	if e != nil && e != io.EOF {
		return
	}

	if read == 0 {
		return io.EOF
	}

	var size int
	if size, e = getBufferSize(buff); e != nil {
		return
	}

	var expectedChecksum [checksumHeaderLen]byte
	copy(expectedChecksum[:], buff[sizeHeaderLen:sizeHeaderLen+checksumHeaderLen])

	out.Write(buff[headerLen:read])

	for out.Len() < size {
		fmt.Println("read xxx")
		read, e = w.conn.Read(buff)
		if e != nil && e != io.EOF {
			return
		}

		out.Write(buff[0:read])

		if e == io.EOF {
			break
		}

	}

	actualChecksum := md5.Sum(out.Bytes())

	if !bytes.Equal(expectedChecksum[0:], actualChecksum[0:]) {
		e = ErrInvalidChecksum
	}

	return
}

// RSAReaderWriter reads and writes data that is encrypted using RSA
type RSAReaderWriter struct {
	publicKey    *rsa.PublicKey
	privateKey   *rsa.PrivateKey
	readerWriter Conn
}

// NewRSAReaderWriter creates RSAReaderWriter
// publicKey - key of the entity I am sending message to
// privateKey - my privateKey
func NewRSAReaderWriter(publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, conn Conn) (rw *RSAReaderWriter) {
	return &RSAReaderWriter{
		publicKey:    publicKey,
		privateKey:   privateKey,
		readerWriter: conn,
	}
}

func (rw *RSAReaderWriter) Read(buff *bytes.Buffer) (e error) {
	var encrypted bytes.Buffer
	if e = rw.readerWriter.Read(&encrypted); e != nil {
		return
	}
	var decrypted []byte
	if decrypted, e = crypto.DecryptOAEP(rw.privateKey, encrypted.Bytes()); e != nil {
		return
	}

	buff.Write(decrypted)
	return
}

func (rw *RSAReaderWriter) Write(buff []byte) (n int, e error) {
	var encrypted []byte
	if encrypted, e = crypto.EncryptOAEP(rw.publicKey, buff); e != nil {
		return
	}

	return rw.readerWriter.Write(encrypted)
}

// CryptoReaderWriter writes and reads bytes from network.  Bytes are
// encrypted on the wire.
type CryptoReaderWriter struct {
	readerWriter         Conn
	block                cipher.Block
	initializationVector []byte
}

func NewCryptoReaderWriter(block cipher.Block, initializationVector []byte, conn Conn) (rw *CryptoReaderWriter) {
	return &CryptoReaderWriter{
		readerWriter:         conn,
		initializationVector: initializationVector,
		block:                block,
	}
}

// Read reads AES encrypted bytes from network, decrypted bytes are
// returned in buff.
func (crw *CryptoReaderWriter) Read(buff *bytes.Buffer) (e error) {
	var encrypted bytes.Buffer
	if e = crw.readerWriter.Read(&encrypted); e != nil {
		return
	}

	decrypted := crypto.DecryptAES(crw.block, crw.initializationVector, encrypted.Bytes())
	buff.Write(decrypted)
	return
}

// Write writes and AES encrypts buff to network
func (crw *CryptoReaderWriter) Write(buff []byte) (n int, e error) {
	encrypted := crypto.EncryptAES(crw.block, crw.initializationVector, buff)
	return crw.readerWriter.Write(encrypted)
}

// GobEncoderReaderWriter reads or writes Go typed data from
// wire
type GobEncoderReaderWriter struct {
	readerWriter Conn
}

func NewGobEncoderReaderWriter(conn Conn) (rw *GobEncoderReaderWriter) {
	return &GobEncoderReaderWriter{
		readerWriter: conn,
	}
}

// Write writes Go types to network
func (g *GobEncoderReaderWriter) Write(v interface{}) (e error) {
	var w bytes.Buffer
	encoder := gob.NewEncoder(&w)
	if e = encoder.Encode(v); e != nil {
		return
	}

	var written int
	if written, e = g.readerWriter.Write(w.Bytes()); e != nil {
		return
	}

	if written != w.Len() {
		e = ErrIncompleteWrite
	}

	return
}

// Read reads Go types from network
func (g *GobEncoderReaderWriter) Read(v interface{}) (e error) {
	var r bytes.Buffer
	if e = g.readerWriter.Read(&r); e != nil {
		return
	}

	decoder := gob.NewDecoder(&r)
	if e = decoder.Decode(v); e != nil {
		return
	}

	return
}
