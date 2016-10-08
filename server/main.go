package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

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

var ErrClientAESKeyAck = errors.New("Client didn't acknowledge receipt of AES keys")

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

	var agent *user.User
	agent, err = handleUserAuthorization(async, s, clientPublicKey)
	if err != nil {
		log.Println("Problem with user authorization: ", err)
		return
	}

	// use AES encryption from here on out
	var aesConn unet.EncodeConn
	aesConn, err = createAESEncryptedConnection(conn, async)
	if err != nil {
		log.Println("Failed to set up AES encrypted connection ", err.Error())
		return
	}

	if err = handleTransfer(agent, aesConn); err != nil {
		log.Println("File transfer failed. ", err.Error())
		return
	}

}

func handleTransfer(agent *user.User, conn unet.EncodeConn) (e error) {
	if e = conn.Write(wire.FileTransferInformationRequest); e != nil {
		return
	}

	var transferInfo wire.FileTransferInformationResponse
	if e = conn.Read(&transferInfo); e != nil {
		return
	}

	if transferInfo.FileTransferType == wire.FileSend {
		e = sendFileToRemote(agent, conn, transferInfo)
	} else {
		e = receiveFileFromRemote(agent, conn, transferInfo)
	}

	return
}

func getIdsFromUser(u *user.User) (uid, gid uint32) {
	val, _ := strconv.ParseUint(u.Uid, 10, 32)
	uid = uint32(val)
	val, _ = strconv.ParseUint(u.Gid, 10, 32)
	gid = uint32(val)
	return
}

// Start process that will read a file as a user (agent) and send contents to stdout, this, the parent process
// reads file bytes from stdout and sends them to remote client
func sendFileToRemote(agent *user.User, conn unet.EncodeConn, transferInfo wire.FileTransferInformationResponse) (e error) {

	uid, gid := getIdsFromUser(agent)

	cmd := exec.Command("ucp_file_reader", fmt.Sprintf("-target-file=%s", transferInfo.FileName))
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}

	var stdOut, stdErr io.ReadCloser
	if stdOut, e = cmd.StdoutPipe(); e != nil {
		return
	}

	if stdErr, e = cmd.StderrPipe(); e != nil {
		return
	}

	if e = cmd.Start(); e != nil {
		return
	}

	errs := make(chan error, 2)

	go func(errs chan<- error) {
		close(errs)
		stdoutErrs := make(chan error)
		processErrs := make(chan error)

		// read from bytes send from child process and send them to remote
		go func(stdout io.ReadCloser, conn unet.EncodeConn, errs chan<- error) {
			defer close(errs)
			var fileStatus wire.FileTransferInformationResponse
			buff := make([]byte, server.PipeBufferSize)
			var read int
			var err error

			if read, err = stdout.Read(buff); err != nil {
				errs <- err
				return
			}

			reader := bytes.NewBuffer(buff[:read])
			decoder := gob.NewDecoder(reader)

			if err = decoder.Decode(&fileStatus); err != nil {
				errs <- err
				return
			}

			for totalRead := int64(0); totalRead < fileStatus.FileSize; {
				if read, err = stdout.Read(buff); err != nil {
					errs <- err
					return
				}

				if err = conn.Write(buff[:read]); err != nil {
					errs <- err
					return
				}

				totalRead += int64(read)
			}

		}(stdOut, conn, stdoutErrs)

		// read problems from child process
		go func(stderr io.ReadCloser, errs chan<- error) {
			defer close(errs)
			buff := make([]byte, server.PipeBufferSize)
			var read int
			var err error
			if read, err = stderr.Read(buff); err != nil {
				errs <- err
				return
			}
			reader := bytes.NewBuffer(buff[:read])
			decoder := gob.NewDecoder(reader)
			var processError error
			if err = decoder.Decode(&processError); err != nil {
				errs <- err
				return
			}

			errs <- processError

		}(stdErr, processErrs)

		// multiplex errors
		select {
		case er, ok := <-stdoutErrs:
			if ok {
				errs <- er
			}
		case er, ok := <-processErrs:
			if ok {
				errs <- er
			}
		}

	}(errs)

	cmd.Wait()

	for err := range errs {
		log.Println("ERROR SendFileToRemote ", err.Error())
	}

	return

}

// Read file bytes from remote client, send bytes to process running under account of user that owns the file via stdin
func receiveFileFromRemote(agent *user.User, conn unet.EncodeConn, transferInfo wire.FileTransferInformationResponse) (e error) {
	uid, gid := getIdsFromUser(agent)
	cmd := exec.Command("ucp_file_writer", fmt.Sprintf("-target-file=%s", transferInfo.FileName))
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}

	var stdIn io.WriteCloser
	if stdIn, e = cmd.StdinPipe(); e != nil {
		return
	}

	if e = cmd.Start(); e != nil {
		return
	}

	for {
		var buffer []byte

		if e = conn.Read(&buffer); e != nil {
			cmd.Process.Kill()
			break
		}

		if len(buffer) > 0 {
			if _, e = stdIn.Write(buffer); e != nil {
				break
			}
		}

		if len(buffer) < server.PipeBufferSize {
			break
		}
	}

	err := cmd.Wait()
	if e == nil && err != nil {
		e = err
	}
	return
}

func createEncryptedConnection(privateKey *rsa.PrivateKey, conn io.ReadWriteCloser) (econn *unet.GobEncoderReaderWriter, clientPubKey *rsa.PublicKey, e error) {
	readerWriter := unet.NewReaderWriter(conn)
	rw := unet.NewGobEncoderReaderWriter(readerWriter)

	// we send public key in plain text to client,
	// client encrypts their key and sends it back to us
	if e = rw.Write(privateKey.PublicKey); e != nil {
		return
	}

	var reader bytes.Buffer
	if e = readerWriter.Read(&reader); e != nil {
		return
	}

	var decrypted []byte
	if decrypted, e = crypto.DecryptOAEP(privateKey, reader.Bytes()); e != nil {
		return
	}

	reader.Reset()
	reader.Write(decrypted)

	clientPubKey = &rsa.PublicKey{}
	decoder := gob.NewDecoder(&reader)
	if e = decoder.Decode(clientPubKey); e != nil {
		return
	}

	econn = unet.NewGobEncoderReaderWriter(
		unet.NewRSAReaderWriter(clientPubKey, privateKey, readerWriter),
	)

	return
}

func createAESEncryptedConnection(rootConn io.ReadWriteCloser, asymmConn unet.EncodeConn) (aesConn unet.EncodeConn, e error) {
	aesParams := wire.SymmetricEncryptionParms{}

	if aesParams.Block, e = crypto.NewCipherBlock(); e != nil {
		return
	}

	aesParams.InitializationVector = make([]byte, crypto.IVBlockSize)
	rand.Read(aesParams.InitializationVector)

	if e = asymmConn.Write(aesParams); e != nil {
		return
	}

	// expect a client response over aes channel
	aesConn = unet.NewGobEncoderReaderWriter(
		unet.NewCryptoReaderWriter(aesParams.Block, aesParams.InitializationVector,
			unet.NewReaderWriter(rootConn),
		),
	)

	if e = aesConn.Read(&aesParams); e != nil {
		return
	}

	if !aesParams.ClientAck {
		e = ErrClientAESKeyAck
	}

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
		conn.Write(authResponse)
		return
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

	if authResponse.AuthResponse == wire.PasswordRequired {
		e = checkUserPassword(conn, s, u)
	}

	return
}

func checkUserPassword(conn unet.EncodeConn, s servicable, user *user.User) (e error) {
	var password string
	if e = conn.Read(&password); e != nil {
		return
	}

	e = s.validatePassword(user, password)

	if e == nil {
		conn.Write(wire.UserAuthorizationResponse{
			AuthResponse: wire.Authorized,
			Description:  "Success",
		})
	} else {
		conn.Write(wire.UserAuthorizationResponse{
			AuthResponse: wire.IncorrectPassword,
			Description:  e.Error(),
		})
	}

	return

}
