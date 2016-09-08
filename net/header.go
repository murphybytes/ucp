package net

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
)

const maxBufferLen = 0x7FFFFFFF

const sizeHeaderLen = 8
const checksumHeaderLen = md5.Size
const headerLen = sizeHeaderLen + checksumHeaderLen

func prependHeaderToBuffer(buffer []byte) (b []byte, e error) {
	var sizeHeader []byte
	if sizeHeader, e = createSizeHeader(buffer); e != nil {
		return
	}

	checkSum := createChecksum(buffer)

	b = make([]byte, 0, len(buffer)+sizeHeaderLen+md5.Size)
	b = append(b, sizeHeader...)
	b = append(b, checkSum[0:]...)
	b = append(b, buffer...)

	return
}

func createSizeHeader(buffer []byte) (h []byte, e error) {
	var size uint64
	size = uint64(len(buffer))
	var buff bytes.Buffer
	if e = binary.Write(&buff, binary.LittleEndian, size); e != nil {
		return
	}
	h = buff.Bytes()

	return
}

func createChecksum(buffer []byte) (s [md5.Size]byte) {
	s = md5.Sum(buffer)
	return

}

func getBufferSize(buffer []byte) (s int, e error) {
	if len(buffer) < headerLen {
		e = fmt.Errorf("Malformed header")
		return
	}
	buf := bytes.NewReader(buffer[0:sizeHeaderLen])
	var i uint32
	if e = binary.Read(buf, binary.LittleEndian, &i); e != nil {
		return
	}

	s = int(i)

	return
}

func getDataPacket(buffer []byte) (data []byte, e error) {
	if len(buffer) < headerLen {
		e = fmt.Errorf("Malformed header")
		return
	}

	data = buffer[headerLen:]
	return
}
