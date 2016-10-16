#!/bin/bash

function test_passed
{
    echo -e "\033[40m\033[1;32m ${test_name} ...Test Passed!  \033[0m"
}

function test_failed
{
    echo -e "\033[40m\033[1;31m ${test_name} ...Test Failed!   \033[0m"
}


function check_result
{
    RESULT=$?
    echo
    if [[ $RESULT == "0" ]]
    then
	test_passed
    else
	test_failed
    fi
}


if [ ! -e "${GOPATH}" ]; then
  echo "Missing GOPATH"
else
  echo "GOPATH = ${GOPATH}"
fi

echo "Setting environments"
export UCP_SERVER_DIRECTORY=$HOME/.ucpserver

echo "Test server key generation"

echo "Y" | userve -generate-keys > /dev/null

check_result

echo "Starting ucp server as a deamon"
echo

userve > $(pwd)/server.log &

echo "Testing urecv generate keys"

echo "Y" | urecv --generate-keys > /dev/null

check_result

echo "Generating test file"
dd if=/dev/urandom of=testfile bs=10000 count=100


echo "Running recv test"
urecv -remote-file=$(pwd)/testfile -local-file=localfile -host=127.0.0.1

check_result

echo "Compare received file with original"
if [ "$(md5 -q testfile)" == "$(md5 -q localfile)" ]; then
  test_passed
else
  test_failed
fi

rm -f testfile
rm -f localfile

kill -15 $(lsof -ti udp:8978)
