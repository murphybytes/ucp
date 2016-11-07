package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	unet "github.com/murphybytes/ucp/net"
	"github.com/murphybytes/ucp/server"
	"github.com/murphybytes/ucp/wire"
)

func main() {
	fmt.Println("child started")
	var socketPath string
	flag.StringVar(&socketPath, "socket-path", "", "Path to unix socket")
	flag.Parse()

	var conn net.Conn
	var err error
	if conn, err = net.Dial("unix", socketPath); err != nil {
		os.Exit(server.ErrSocket)
	}

	if err = handleConnection(conn); err != nil {
		os.Exit(server.Error)
	}

	os.Exit(server.Success)

}

func handleConnection(conn net.Conn) (e error) {
	defer conn.Close()
	rw := unet.NewReaderWriter(conn)
	encoderConn := unet.NewGobEncoderReaderWriter(rw)

	var transferInfo wire.FileTransferInformation

	if e = encoderConn.Read(&transferInfo); e != nil {
		return
	}

	fmt.Printf("child read txfer %+v\n", transferInfo)

	if transferInfo.FileTransferType == wire.FileSend {
		f := newOsFile()
		return fileSend(encoderConn, transferInfo, f)
	}

	// TODO: handle recieving file from parent process

	return
}
