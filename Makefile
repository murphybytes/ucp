OS="OSX"
ARCH="AMD64"
UDTDIR="vendor/github.com/murphybytes/udt.go/vendor/udt"

build_server:
	go build -o userve github.com/murphybytes/ucp/server

build_send:
	go build -o usend github.com/murphybytes/ucp/send; mv usend $(GOPATH)/bin

test_send:
	go test -v github.com/murphybytes/ucp/send

test_recv:
	go test -v github.com/murphybytes/ucp/recv

test_client:
	go test -v github.com/murphybytes/ucp/client

build_recv:
	go build -o urecv github.com/murphybytes/ucp/recv; mv urecv $(GOPATH)/bin

build_udt:
	cd $(UDTDIR); make clean; make -e os=$(OS) arch=$(ARCH);cp src/libudt.* $(GOPATH)/bin/.; make clean

test_net:
	go test -v github.com/murphybytes/ucp/net

test_crypto:
	go test -v github.com/murphybytes/ucp/crypto

test: test_net test_crypto test_send test_recv test_client

all: build_udt build_server build_recv build_send

.PHONY: build_udt build_server all test test_net test_crypto test_send test_recv test_client
