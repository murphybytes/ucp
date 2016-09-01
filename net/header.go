package net

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
)

const sizeHeaderSize = 4
const headerSize = sizeHeaderSize + md5.Size

func buildPacket(buffer []byte) (b []byte, e error) {
	var sizeHeader []byte
	if sizeHeader, e = getSizeHeader(buffer); e != nil {
		return
	}

	checkSum := getChecksum(buffer)

	b = make([]byte, 0, len(buffer)+sizeHeaderSize+md5.Size)
	b = append(b, sizeHeader...)
	b = append(b, checkSum[0:]...)
	b = append(b, buffer...)

	return
}

func getSizeHeader(buffer []byte) (h []byte, e error) {
	var size uint32
	size = uint32(len(buffer))
	var buff bytes.Buffer
	if e = binary.Write(&buff, binary.LittleEndian, size); e != nil {
		return
	}
	h = buff.Bytes()

	return
}

func getChecksum(buffer []byte) (s [md5.Size]byte) {
	s = md5.Sum(buffer)
	return

}

func compareChecksum(packet []byte) (e error) {
	if len(packet) < headerSize {
		return fmt.Errorf("Malformed packet")
	}

	checksum := md5.Sum(packet[headerSize:])

	comp := bytes.Compare(checksum[0:], packet[4:4+md5.Size])
	if comp != 0 {
		e = fmt.Errorf("Checksum comparison failed")
		return
	}

	return
}

func getBufferSize(buffer []byte) (s int, e error) {
	if len(buffer) < headerSize {
		e = fmt.Errorf("Malformed header")
		return
	}
	buf := bytes.NewReader(buffer[0:sizeHeaderSize])
	var i uint32
	if e = binary.Read(buf, binary.LittleEndian, &i); e != nil {
		return
	}

	s = int(i)

	return
}

func getDataPacket(buffer []byte) (data []byte, e error) {
	if len(buffer) < headerSize {
		e = fmt.Errorf("Malformed header")
		return
	}

	data = buffer[headerSize:]
	return
}
