package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"io"

	"os"

	"github.com/murphybytes/ucp/server/shared"
	"github.com/murphybytes/ucp/wire"
)

// Program reads a file and writes contents to stdout where it will be read in server and sent somewhere
func main() {

	var targetFile string
	flag.StringVar(&targetFile, "target-file", "", "name of the file that will be read from.")

	if file, err := os.Open(targetFile); err == nil {
		defer file.Close()

		if err = sendFileSizeToParentProcess(file); err != nil {
			sendErrorToParentProcess(err)
		}

		buffer := make([]byte, server.PipeBufferSize)

		for {
			var read int
			read, err = file.Read(buffer)

			if read > 0 {
				os.Stdout.Write(buffer[:read])
			}

			if read < server.PipeBufferSize || err == io.EOF {
				os.Exit(server.Success)
			}

			if err != nil {
				os.Exit(server.Error)
			}
		}

	} else {
		sendErrorToParentProcess(err)
	}
}

func sendErrorToParentProcess(e error) {
	var writer bytes.Buffer
	encoder := gob.NewEncoder(&writer)
	encoder.Encode(e)
	os.Stderr.Write(writer.Bytes())
	os.Exit(server.Error)
}

func sendFileSizeToParentProcess(file *os.File) (e error) {
	var txferInfo wire.FileTransferInformationResponse
	var fileInfo os.FileInfo
	// tell parent process how many bytes we'll be sending
	if fileInfo, e = file.Stat(); e != nil {
		return
	}

	txferInfo.FileSize = fileInfo.Size()

	var gobBuffer bytes.Buffer
	encoder := gob.NewEncoder(&gobBuffer)

	if e = encoder.Encode(txferInfo); e != nil {
		return
	}

	os.Stdout.Write(gobBuffer.Bytes())

	return
}
