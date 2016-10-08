package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"

	"os"

	"github.com/murphybytes/ucp/server/shared"
)

// This program is designed to be kicked off by the server process, su'ed to
// a different user, read file bytes over stdin and write them to a file
func main() {

	var targetFile string

	flag.StringVar(&targetFile, "target-file", "", "file to write to")

	if file, err := os.Create(targetFile); err == nil {
		defer file.Close()

		reader := bufio.NewReader(os.Stdin)
		buffer := make([]byte, server.PipeBufferSize)

		for {
			read, err := reader.Read(buffer)

			if err != nil && err != io.EOF {
				fmt.Println("Error reading from server", err.Error())
				os.Exit(server.Error)
			}

			_, erroutfile := file.Write(buffer[:read])
			if erroutfile != nil {
				fmt.Println("Error writing out file", erroutfile.Error())
				os.Exit(server.Error)
			}

			if err == io.EOF || read < server.PipeBufferSize {
				os.Exit(server.Success)
			}
		}

	}

	fmt.Println("Could not create file ", targetFile)
	os.Exit(server.Error)

}
