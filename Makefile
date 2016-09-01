OS="OSX"
ARCH="AMD64"
UDTDIR="vendor/udt/udt4"

build_server:
	go build -o userve github.com/murphybytes/ucp/server; mv userve $(GOPATH)/bin

build_send:
	go build -o usend github.com/murphybytes/ucp/send; mv usend $(GOPATH)/bin

build_recv:
	go build -o urecv github.com/murphybytes/ucp/recv; mv urecv $(GOPATH)/bin

build_udt:
	git submodule update --init; cd $(UDTDIR); make clean; make -e os=$(OS) arch=$(ARCH)

all: build_udt build_server build_recv build_send 

.PHONY: build_udt build_server
