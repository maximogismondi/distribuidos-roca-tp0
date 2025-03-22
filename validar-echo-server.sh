#!/bin/bash
echo "This script will validate the echo server"

TEST_TEXT="Hello, World!"
REGEX="^$TEST_TEXT\$"

# ejecutar un alpine básico en la red testing_net

docker run --rm --network tp0_testing_net alpine sh -c "\
    echo \"$TEST_TEXT\" | nc server 12345 | grep -q \"$REGEX\""

# ahora en funcion del resultado imprimir un mensaje adecuado

if [ $? -eq 0 ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi