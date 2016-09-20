package client

import (
	"crypto/rsa"
	"flag"
	"net"
	"os"
	"os/user"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	unet "github.com/murphybytes/ucp/net"
	"github.com/murphybytes/ucp/server/shared"
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

func CreateEncryptedConnection(privateKey *rsa.PrivateKey, conn net.Conn) (econn *unet.GobEncoderReaderWriter, e error) {
	readerWriter := unet.NewReaderWriter(conn)
	rw := unet.NewGobEncoderReaderWriter(readerWriter)

	var serverPublicKey rsa.PublicKey
	if e = rw.Read(&serverPublicKey); e != nil {
		return
	}

	if e = rw.Write(privateKey.PublicKey); e != nil {
		return
	}

	econn = unet.NewGobEncoderReaderWriter(
		unet.NewRSAReaderWriter(&serverPublicKey, privateKey, readerWriter),
	)

	return
}
