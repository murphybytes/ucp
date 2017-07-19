package udt

import (
	"github.com/murphybytes/ucp/udt/cudt"
)

// Startup sets up resources used by UDT.
func Startup() error {
	return cudt.Startup()
}

// Cleanup frees resources used by UDT.
func Cleanup() error {
	return cudt.Cleanup()
}
