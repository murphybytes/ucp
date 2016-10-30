package main

import (
	"flag"
	"io"

	"os"

	"github.com/murphybytes/ucp/net"
	"github.com/murphybytes/ucp/server/shared"
	"github.com/murphybytes/ucp/wire"
)

// Program reads a file and writes contents to stdout where it will be read in server and sent somewhere
func main() {

	var targetFile string
	flag.StringVar(&targetFile, "target-file", "", "name of the file that will be read from.")
	flag.Parse()

	readerWriter := server.NewReadWriteJoiner(os.Stdin, os.Stdout)
	parentConn := net.NewGobEncoderReaderWriter(readerWriter)

	if file, err := os.Open(targetFile); err == nil {
		defer file.Close()

		if err = sendFileSizeToParentProcess(parentConn, file); err != nil {
			os.Exit(server.Error)
		}

		buffer := make([]byte, server.PipeBufferSize)

		for {
			var chunkInfo wire.FileChunk

			var read int
			read, err = file.Read(buffer)

			if err != nil {
				chunkInfo.Error = err
				parentConn.Write(chunkInfo)
				waitOnParentProcess(parentConn)
				os.Exit(server.Error)
			}

			if read > 0 {
				chunkInfo.Buffer = buffer[:read]
				if err = parentConn.Write(chunkInfo); err != nil {
					os.Exit(server.Error)
				}
			}

			if read < server.PipeBufferSize || err == io.EOF {
				// block until server reads and says we're done
				waitOnParentProcess(parentConn)
				os.Exit(server.Success)
			}

		}

	} else {
		//	sendErrorToParentProcess(err)
		transferResponse := wire.FileTransferInformationResponse{
			Error: err,
		}
		parentConn.Write(transferResponse)
		waitOnParentProcess(parentConn)
	}
}

// waitOnParentProcess will block until we get a response from parent
// process. This is so this process doesn't exit and close pipes before
// the parent gets a chance to read from them
func waitOnParentProcess(conn net.EncodeConn) {
	var response wire.Conversation
	conn.Read(&response)
}

func sendFileSizeToParentProcess(conn net.EncodeConn, file *os.File) (e error) {
	var txferInfo wire.FileTransferInformationResponse
	var fileInfo os.FileInfo
	// tell parent process how many bytes we'll be sending
	if fileInfo, e = file.Stat(); e != nil {
		txferInfo.Error = e
		conn.Write(txferInfo)
		waitOnParentProcess(conn)
		return
	}

	txferInfo.FileSize = fileInfo.Size()
	e = conn.Write(txferInfo)

	return
}
