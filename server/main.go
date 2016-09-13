package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/murphybytes/ucp/crypto"
	"github.com/murphybytes/ucp/server/shared"
	"github.com/murphybytes/udt.go/udt"
)

const errorCode = 1
const successCode = 0
const defaultInterface = "localhost"

var generateKeys bool
var ucpDirectory string
var hostInterface string

func init() {

	flag.BoolVar(&generateKeys, "generate-keys", false, "Generate rsa keys and exit.")
	flag.StringVar(&ucpDirectory, "ucp-directory", os.Getenv("UCP_DIRECTORY"), "Directory where keys and other application files are stored")
	flag.StringVar(&hostInterface, "host-interface", fmt.Sprintf("localhost:%d", server.DefaultPort), "Interface that server will listen on")

}

func main() {
	flag.Parse()

	if generateKeys {
		fmt.Println("Creating UCP keys and files in ", ucpDirectory)

		if err := crypto.InitializeUcpDir(ucpDirectory); err != nil {
			fmt.Println(err)
			os.Exit(errorCode)
		}
		os.Exit(successCode)
	}

	if err := udt.Startup(); err != nil {
		log.Printf("Init failed with error %s\n", err.Error())
		os.Exit(errorCode)
	}

	defer udt.Cleanup()

	if listener, err := udt.Listen(hostInterface); err == nil {
		defer listener.Close()

		for {
			var conn net.Conn
			if conn, err = listener.Accept(); err == nil {
				// TODO: handle connection
				handleConnection(conn)
			} else {
				log.Println("Error accepting connection ", err)
			}
		}
	} else {
		log.Println("Error establishing server interface. ", err)
	}

}

func handleConnection(conn net.Conn) {

}
