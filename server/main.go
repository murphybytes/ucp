package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	_ "github.com/joho/godotenv/autoload"
	"github.com/murphybytes/ucp/crypto"
	unet "github.com/murphybytes/ucp/net"
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
	flag.StringVar(&ucpDirectory, "ucp-directory", os.Getenv("UCP_SERVER_DIRECTORY"), "Directory where keys and other application files are stored")
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
				go handleConnection(conn)
			} else {
				log.Println("Error accepting connection ", err)
			}
		}
	} else {
		log.Println("Error establishing server interface. ", err)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	privateKey, err := crypto.GetPrivateKey(filepath.Join(ucpDirectory, "private-key.pem"))
	if err != nil {
		log.Println("Can't fetch private key: ", err)
	}

	var async *unet.GobEncoderReaderWriter
	if async, err = createAsynchEncryptionConnection(privateKey, conn); err != nil {
		log.Println("ERROR: ", err)
	}

	random := make([]byte, 50)
	rand.Read(random)
	if err = async.Write(random); err != nil {
		log.Println("ERROR: Write failed ", err)
		return
	}

	var response []byte
	if err = async.Read(&response); err != nil {
		log.Println("ERROR: Read failed ", err)
		return
	}

	if bytes.Equal(random, response) {
		log.Println("Async key exchange succeeded")
	} else {
		log.Println("ERROR: Async key exchange failed. ")
	}

}

func createAsynchEncryptionConnection(privateKey *rsa.PrivateKey, conn net.Conn) (async *unet.GobEncoderReaderWriter, e error) {
	readerWriter := unet.NewReaderWriter(conn)
	rw := unet.NewGobEncoderReaderWriter(readerWriter)

	if e = rw.Write(privateKey.PublicKey); e != nil {
		return
	}

	var clientPublicKey rsa.PublicKey
	if e = rw.Read(&clientPublicKey); e != nil {
		return
	}

	async = unet.NewGobEncoderReaderWriter(
		unet.NewRSAReaderWriter(&clientPublicKey, privateKey, readerWriter),
	)
	return

}
