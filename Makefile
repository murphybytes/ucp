OS="OSX"
ARCH="AMD64"
UDTDIR="vendor/github.com/murphybytes/udt.go/vendor/udt"
RACE=-race

build_server:
	go install $(RACE) github.com/murphybytes/ucp/userve
	go install github.com/murphybytes/ucp/uproxy


build_send:
	go build -o usend github.com/murphybytes/ucp/send; ln -sf $(shell pwd)/usend $(GOPATH)/bin/.

build_recv:
	go build -o urecv github.com/murphybytes/ucp/recv; ln -sf $(shell pwd)/urecv $(GOPATH)/bin/.


test_send:
	go test -v github.com/murphybytes/ucp/send

test_recv:
	go test -v github.com/murphybytes/ucp/recv

test_client:
	go test -v github.com/murphybytes/ucp/client

test_server:
	go test -v github.com/murphybytes/ucp/server

test_userve:
	go test -v github.com/murphybytes/ucp/userve

build_udt:
	cd $(UDTDIR); make clean; make -e os=$(OS) arch=$(ARCH);cp src/libudt.* $(GOPATH)/bin/.; make clean

test_net:
	go test -v github.com/murphybytes/ucp/net

test_crypto:
	go test -v github.com/murphybytes/ucp/crypto

test_uproxy:
	go test -v github.com/murphybytes/ucp/uproxy

test: test_net test_crypto test_send test_recv test_client test_server test_userve test_uproxy

all: build_udt build_server build_recv build_send

.PHONY: build_udt build_server all test test_net test_crypto test_send test_recv test_client test_server test_userve test_uproxy
