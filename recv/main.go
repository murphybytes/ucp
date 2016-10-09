package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/murphybytes/ucp/client"
	"github.com/murphybytes/ucp/crypto"
	unet "github.com/murphybytes/ucp/net"
	"github.com/murphybytes/ucp/wire"
	"github.com/murphybytes/udt.go/udt"
)

// reads from remote file, writes local
func main() {
	var localFilePath, remoteFilePath string
	flag.StringVar(&localFilePath, "local-file", "", "File where recieved data will be written")
	flag.StringVar(&remoteFilePath, "remote-file", "", "File where data will be read from")
	flag.Parse()

	if client.ShowHelp {
		flag.PrintDefaults()
		os.Exit(client.ErrorCode)
	}

	if client.GenerateKeys {
		fmt.Println("Creating UCP keys and files in ", client.UCPDirectory)

		if err := crypto.InitializeUcpDir(client.UCPDirectory); err != nil {
			fmt.Println(err)
			os.Exit(client.ErrorCode)
		}
		os.Exit(client.SuccessCode)
	}

	var err error
	err = udt.Startup()
	client.ExitOnError(err, "Could not initialize UDT library")
	defer udt.Cleanup()

	networkEndpoint := fmt.Sprintf("%s:%d", client.Host, client.Port)

	var conn net.Conn
	conn, err = udt.Dial(networkEndpoint)
	client.ExitOnError(err, "Could not connect to", networkEndpoint)

	privateKey, err := crypto.GetPrivateKey(filepath.Join(client.UCPDirectory, "private-key.pem"))
	if err != nil {
		fmt.Println(err)
		os.Exit(client.ErrorCode)
	}

	var asymmEncryptedConn *unet.GobEncoderReaderWriter
	asymmEncryptedConn, err = client.CreateRSAEncryptedConnection(privateKey, conn)
	client.ExitOnError(err)

	var prompt client.Prompt
	err = client.HandleUserAuthorization(asymmEncryptedConn, &prompt)
	client.ExitOnError(err, "User authorization failed")

	var aesEncryptedConn unet.EncodeConn
	aesEncryptedConn, err = client.CreateAESEncryptedConnection(conn, asymmEncryptedConn)
	client.ExitOnError(err, "Failed to establish aes encrypted connection")

	err = receiveFileFromServer(localFilePath, remoteFilePath, aesEncryptedConn)
	client.ExitOnError(err, "File transfer failed")

}

func receiveFileFromServer(localPath, remotePath string, conn unet.EncodeConn) (e error) {
	var request wire.Conversation
	if e = conn.Read(&request); e != nil {
		return
	}

	if request != wire.FileTransferInformationRequest {
		return client.ErrBadRequest
	}

	transferInfo := wire.FileTransferInformationResponse{
		FileTransferType: wire.FileSend,
		FileName:         remotePath,
	}

	if e = conn.Write(transferInfo); e != nil {
		return
	}

	if e = conn.Read(&transferInfo); e != nil {
		return
	}

	var localFile *os.File
	if localFile, e = os.Create(localPath); e != nil {
		conn.Write(wire.FileTransferAbort)
		return
	}
	defer localFile.Close()

	if e = conn.Write(wire.FileTransferStart); e != nil {
		return
	}

	for totalRead := int64(0); totalRead < transferInfo.FileSize; {
		var buffer []byte
		if e = conn.Read(&buffer); e != nil {
			return
		}

		totalRead += int64(len(buffer))

		if _, e = localFile.Write(buffer); e != nil {
			conn.Write(wire.FileTransferFail)
			return
		}

		if e = conn.Write(wire.FileTransferMore); e != nil {
			return
		}

	}

	return
}
