package cudt

// #cgo CXXFLAGS: -I${SRCDIR}/../src/udt/src
// #cgo LDFLAGS: -L${SRCDIR}/../src/udt/src -ludt -lstdc++ -lpthread -lm
import "C"
