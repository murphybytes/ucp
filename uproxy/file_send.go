package main

import (
	"io"
	"os"

	unet "github.com/murphybytes/ucp/net"
	"github.com/murphybytes/ucp/server"
	"github.com/murphybytes/ucp/wire"
)

type fileIntf interface {
	open(fileName string) (reader io.ReadCloser, e error)
	getFileSize() (size int64, e error)
}

type osFile struct {
	f *os.File
}

func newOsFile() *osFile {
	return &osFile{}
}

func (o *osFile) open(fileName string) (f io.ReadCloser, e error) {
	o.f, e = os.Open(fileName)
	return o.f, e
}

func (o *osFile) getFileSize() (size int64, e error) {
	var fileInfo os.FileInfo
	fileInfo, e = o.f.Stat()
	size = fileInfo.Size()
	return
}

func fileSend(conn unet.EncodeConn, txferInfo wire.FileTransferInformation, f fileIntf) (e error) {
	var file io.ReadCloser
	if file, e = f.open(txferInfo.FileName); e != nil {
		txferInfo.Error = e
		conn.Write(txferInfo)
		return
	}
	file.Close()

	var fileSize int64
	if fileSize, e = f.getFileSize(); e != nil {
		txferInfo.Error = e
		conn.Write(txferInfo)
		return
	}

	// tell parent process how many bytes we'll be sending
	txferInfo.FileSize = fileSize
	if e = conn.Write(txferInfo); e != nil {
		return
	}

	e = sendFileBytesToParentProcess(conn, file, fileSize)

	return
}

func sendFileBytesToParentProcess(conn unet.EncodeConn, file io.Reader, bytesToSend int64) (e error) {

	readBuffer := make([]byte, server.FileReaderBufferSize)

	for totalRead := int64(0); totalRead < bytesToSend; {
		var read int
		var chunk wire.FileChunk
		if read, e = file.Read(readBuffer); e != nil {
			chunk.Error = e
			conn.Write(chunk)
			return
		}

		totalRead += int64(read)

		chunk.Buffer = readBuffer[:read]

		if e = conn.Write(chunk); e != nil {
			return
		}

		var response wire.Conversation
		if e = conn.Read(&response); e != nil {
			return
		}

		if response != wire.FileTransferMore {
			return server.ErrParentTerminatedConversation
		}

	}

	return
}
