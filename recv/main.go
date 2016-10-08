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
	asymmEncryptedConn, err = client.CreateEncryptedConnection(privateKey, conn)
	client.ExitOnError(err)

	var prompt client.Prompt
	err = client.HandleUserAuthorization(asymmEncryptedConn, prompt)
	client.ExitOnError(err, "User authorization failed")

}
