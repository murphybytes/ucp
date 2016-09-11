package client

import (
	"flag"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	"github.com/murphybytes/ucp/server/shared"
)

// UCPDirectory path to keys and known_hosts file
var UCPDirectory string

// Host ip address or host name of server
var Host string

// Port port that ucp server is listening on defaults to 8978
var Port int

func init() {
	UCPDirectory = os.Getenv("UCP_DIRECTORY")

	flag.StringVar(&Host, "host", os.Getenv("UCP_HOST"), "IP Address or Hostname for UCP server")
	flag.IntVar(&Port, "port", getIntFromEnvironment(os.Getenv("UCP_PORT"), server.DefaultPort), "Port for UCP server")
}

func getIntFromEnvironment(envVal string, defaultVal int) (r int) {
	var err error
	if r, err = strconv.Atoi(envVal); err != nil {
		r = defaultVal
	}
	return
}
