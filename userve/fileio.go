package main

import (
	"fmt"

	"github.com/murphybytes/ucp/net"
	"github.com/murphybytes/ucp/wire"
)

func readFromChildProcessAndSendToRemote(childProcessConn net.EncodeConn, remoteConn net.EncodeConn, transferInfo wire.FileTransferInformation) (e error) {
	defer childProcessConn.Write(wire.FileTransferComplete)
	fmt.Printf("read from child process %+v\n", transferInfo)

	// Tell child process we want to send file
	if e = childProcessConn.Write(&transferInfo); e != nil {
		return
	}

	// Read file size from child process
	if e = childProcessConn.Read(&transferInfo); e != nil {
		return
	}

	fmt.Printf("Got transfer info %+v\n", transferInfo)

	if transferInfo.Error != nil {
		return transferInfo.Error
	}

	// Send file size to remote client so it will know how many bytes to expect
	if e = remoteConn.Write(transferInfo); e != nil {
		return
	}

	var remoteClientMessage wire.Conversation
	if e = remoteConn.Read(&remoteClientMessage); e != nil {
		return
	}

	fmt.Println("Remote is ready to get data")

	if remoteClientMessage != wire.FileTransferStart {
		return ErrClientFileTxferAbort
	}

	for totalRead := int64(0); totalRead < transferInfo.FileSize; {
		fmt.Println("Read loop")
		var chunk wire.FileChunk

		if e = childProcessConn.Read(&chunk); e != nil {
			return
		}

		fmt.Printf("Read chunk Err %+v\n", chunk.Error)

		if chunk.Error != nil {
			return chunk.Error
		}

		totalRead += int64(len(chunk.Buffer))

		fmt.Printf("Read %d of %d\n", totalRead, transferInfo.FileSize)

		if e = remoteConn.Write(chunk.Buffer); e != nil {
			return
		}

		if e = remoteConn.Read(&remoteClientMessage); e != nil {
			return
		}

		if remoteClientMessage != wire.FileTransferMore {
			return fmt.Errorf("Connection prematurely terminated by remote client")
		}

		fmt.Println("remote client requests more")

		if e = childProcessConn.Write(wire.FileTransferMore); e != nil {
			// TODO: tell remote we're fucked
			return
		}

	}

	return
}
