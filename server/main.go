package main

import (
	"crypto/rsa"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"

	_ "github.com/joho/godotenv/autoload"
	"github.com/murphybytes/ucp/crypto"
	unet "github.com/murphybytes/ucp/net"
	"github.com/murphybytes/ucp/server/shared"
	"github.com/murphybytes/ucp/wire"
	"github.com/murphybytes/udt.go/udt"
	"golang.org/x/crypto/ssh"
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
	var err error
	flag.Parse()

	if generateKeys {
		fmt.Println("Creating UCP keys and files in ", ucpDirectory)

		if err = crypto.InitializeUcpDir(ucpDirectory); err != nil {
			fmt.Println(err)
			os.Exit(errorCode)
		}
		os.Exit(successCode)
	}

	if err = udt.Startup(); err != nil {
		log.Printf("Init failed with error %s\n", err.Error())
		os.Exit(errorCode)
	}

	defer udt.Cleanup()

	var service *osService
	if service, err = newOsService(); err != nil {
		log.Println("Service initialization failed: ", err)
		os.Exit(errorCode)
	}

	if listener, err := udt.Listen(hostInterface); err == nil {
		defer listener.Close()

		for {
			var conn net.Conn
			if conn, err = listener.Accept(); err == nil {
				// TODO: handle connection
				go handleConnection(conn, service)
			} else {
				log.Println("Error accepting connection ", err)
			}
		}
	} else {
		log.Println("Error establishing server interface. ", err)
	}

}

func handleConnection(conn net.Conn, s servicable) {
	defer conn.Close()
	privateKey := s.getPrivateKey()
	var err error
	var async *unet.GobEncoderReaderWriter
	var clientPublicKey *rsa.PublicKey

	if async, clientPublicKey, err = createEncryptedConnection(privateKey, conn); err != nil {
		log.Println("ERROR: ", err)
	}

	_, err = handleUserAuthorization(async, s, clientPublicKey)
	if err != nil {
		log.Println("Problem with user authorization: ", err)
		return
	}

}

func createEncryptedConnection(privateKey *rsa.PrivateKey, conn io.ReadWriteCloser) (econn *unet.GobEncoderReaderWriter, clientPubKey *rsa.PublicKey, e error) {
	readerWriter := unet.NewReaderWriter(conn)
	rw := unet.NewGobEncoderReaderWriter(readerWriter)

	if e = rw.Write(privateKey.PublicKey); e != nil {
		return
	}

	clientPubKey = &rsa.PublicKey{}
	if e = rw.Read(clientPubKey); e != nil {
		return
	}

	econn = unet.NewGobEncoderReaderWriter(
		unet.NewRSAReaderWriter(clientPubKey, privateKey, readerWriter),
	)

	return
}

func handleUserAuthorization(conn unet.EncodeConn, s servicable, clientPubKey *rsa.PublicKey) (u *user.User, e error) {

	if e = conn.Write(wire.UserNameRequest); e != nil {
		return
	}

	var userName string
	if e = conn.Read(&userName); e != nil {
		return
	}

	if u, e = s.lookupUser(userName); e != nil {
		authResponse := wire.UserAuthorizationResponse{
			AuthResponse: wire.NonexistantUser,
			Description:  fmt.Sprintf("User '%s' is unknown", userName),
		}
		if e = conn.Write(authResponse); e != nil {
			return
		}
	}

	// got user see if requestor's public key in in authorized keys
	var sshPublicKey ssh.PublicKey
	if sshPublicKey, e = ssh.NewPublicKey(clientPubKey); e != nil {
		return
	}

	encodedPublicKey := ssh.MarshalAuthorizedKey(sshPublicKey)

	var keyinAuthorizedKeys bool
	keyinAuthorizedKeys, e = s.isKeyAuthorized(u, encodedPublicKey,
		func() []byte {
			if reader, err := os.Open(filepath.Join(u.HomeDir, ".ucp", "authorized_keys")); err == nil {
				defer reader.Close()
				if contents, ee := ioutil.ReadAll(reader); ee == nil {
					return contents
				}
			}
			return []byte{}
		})

	if e != nil {
		return
	}

	authResponse := wire.UserAuthorizationResponse{}

	if keyinAuthorizedKeys {
		authResponse.AuthResponse = wire.Authorized
	} else {
		authResponse.AuthResponse = wire.PasswordRequired
	}

	if e = conn.Write(authResponse); e != nil {
		return
	}

	// if authResponse.AuthResponse == wire.PasswordRequired {
	// 	e = checkUserPassword(conn, s, u)
	// }

	return
}

func checkUserPassword(conn unet.EncodeConn, s servicable, user *user.User) (e error) {
	var password string
	if e = conn.Read(&password); e != nil {
		return
	}
	return s.validatePassword(user, password)

}
