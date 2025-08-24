#!/bin/bash

echo "Testing echo server"

EXPECTED="They’re the same picture"
SERVER_ADDR="server:12345"

result=$(docker run --rm --network tp0_testing_net busybox sh -c "echo $EXPECTED | nc $SERVER_ADDR")

if [ "$result" == "$EXPECTED" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
