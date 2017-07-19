

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
		SYSTEM = LINUX
		LIB_ENV_VAR=LDD_LIBRARY_PATH=$(shell pwd)/src/udt/src
endif
ifeq ($(UNAME_S),Darwin)
		SYSTEM = OSX
		LIB_ENV_VAR=DYLD_LIBRARY_PATH=$(shell pwd)/src/udt/src
endif

UNAME_P := $(shell uname -p)
ifeq ($(UNAME_P),x86_64)
		PLATFORM = AMD64
endif
ifneq ($(filter %86,$(UNAME_P)),)
		PLATFORM = IA32
endif

test: udt
	$(LIB_ENV_VAR) go test github.com/murphybytes/ucp/udt/cudt

udt:
	$(MAKE) -C src/udt -e os=$(SYSTEM) arch=$(PLATFORM)

build: udt
	go build github.com/murphybytes/ucp/udt

.PHONY: udt test build 
