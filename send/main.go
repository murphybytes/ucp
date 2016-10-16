package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/murphybytes/ucp/client"
	"github.com/murphybytes/ucp/crypto"
	"github.com/murphybytes/udt.go/udt"
)

func main() {
	fmt.Println("Starting")
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

	if err := udt.Startup(); err != nil {
		fmt.Println("Unable to initialize UDT Library: ", err)
		os.Exit(client.ErrorCode)
	}
	defer udt.Cleanup()

	connectStr := fmt.Sprintf("%s:%d", client.Host, client.Port)
	var conn net.Conn
	if c, err := udt.Dial(connectStr); err != nil {
		fmt.Println(err)
		os.Exit(client.ErrorCode)
	} else {
		conn = c
	}

	privateKey, err := crypto.GetPrivateKey(filepath.Join(client.UCPDirectory, "private-key.pem"))
	if err != nil {
		fmt.Println(err)
		os.Exit(client.ErrorCode)
	}

	asyncConn, err := client.CreateRSAEncryptedConnection(privateKey, conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(client.ErrorCode)
	}

	if err = client.HandleUserAuthorization(asyncConn, &client.Prompt{}); err != nil {
		fmt.Println("User authorization failed: ", err)
		os.Exit(client.ErrorCode)
	}

	os.Exit(client.SuccessCode)

}
