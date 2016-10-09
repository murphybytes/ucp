package client

import (
	"crypto/rsa"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/user"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	unet "github.com/murphybytes/ucp/net"
	"github.com/murphybytes/ucp/server/shared"
	"github.com/murphybytes/ucp/wire"
)

// TODO: Add option to compress files before send/recv

const (
	ErrorCode   = 1
	SuccessCode = 0
)

// UCPDirectory path to keys and known_hosts file
var UCPDirectory string

// Host ip address or host name of server
var Host string

// Port port that ucp server is listening on defaults to 8978
var Port int

var GenerateKeys bool
var ShowHelp bool

var RemoteUser string

var ErrBadRequest = errors.New("Unexpected or invalid request")

func init() {
	UCPDirectory = os.Getenv("UCP_DIRECTORY")

	flag.StringVar(&Host, "host", os.Getenv("UCP_HOST"), "IP Address or Hostname for UCP server")
	flag.StringVar(&RemoteUser, "user", GetCurrentUserName(), "The name of the user who owns the file")
	flag.IntVar(&Port, "port", getIntFromEnvironment(os.Getenv("UCP_PORT"), server.DefaultPort), "Port for UCP server")
	flag.BoolVar(&GenerateKeys, "generate-keys", false, "Generate rsa keys and exit.")
	flag.BoolVar(&ShowHelp, "help", false, "Show help message.")
}

func getIntFromEnvironment(envVal string, defaultVal int) (r int) {
	var err error
	if r, err = strconv.Atoi(envVal); err != nil {
		r = defaultVal
	}
	return
}

func GetCurrentUserName() string {
	if u, e := user.Current(); e == nil {
		return u.Username
	}
	return ""
}

func ExitOnError(e error, msgs ...string) {
	if e != nil {
		descriptions := ""
		for _, msg := range msgs {
			if descriptions != "" {
				descriptions += " "
			}
			descriptions += msg
		}

		fmt.Println(descriptions, e.Error())
		os.Exit(ErrorCode)
	}
}

// CreateRSAEncryptedConnection creates a network connection that will RSA encrypt
// bytes before sending them.  Takes an RSA private key and a network connection
// as arguments.
func CreateRSAEncryptedConnection(privateKey *rsa.PrivateKey, conn net.Conn) (econn *unet.GobEncoderReaderWriter, e error) {
	readerWriter := unet.NewReaderWriter(conn)
	rw := unet.NewGobEncoderReaderWriter(readerWriter)

	var serverPublicKey rsa.PublicKey
	if e = rw.Read(&serverPublicKey); e != nil {
		return
	}

	econn = unet.NewGobEncoderReaderWriter(
		unet.NewRSAReaderWriter(&serverPublicKey, privateKey, readerWriter),
	)

	if e = econn.Write(privateKey.PublicKey); e != nil {
		return
	}

	return
}

// CreateAESEncryptedConnection creates a connection that uses AES encryption.  AES (Symmetric Key) encryption is much faster than RSA
func CreateAESEncryptedConnection(rootConn net.Conn, asyncEncryptedConn unet.EncodeConn) (aesEncryptedConn unet.EncodeConn, e error) {
	var aesParams wire.SymmetricEncryptionParms
	if e = asyncEncryptedConn.Read(&aesParams); e != nil {
		return
	}

	aesEncryptedConn = unet.NewGobEncoderReaderWriter(
		unet.NewCryptoReaderWriter(aesParams.Block, aesParams.InitializationVector,
			unet.NewReaderWriter(rootConn)))

	// ack that we've established aes encrypted connection
	aesParams.ClientAck = true
	e = aesEncryptedConn.Write(aesParams)

	return
}
