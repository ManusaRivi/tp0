#!/bin/sh

cat <<EOF > Dockerfile.Echotest
FROM alpine:latest
RUN apk add --no-cache netcat-openbsd
EOF

docker build -f Dockerfile.Echotest -t echo-test-client .

result=`echo "Test" | docker run -i --rm --network tp0_testing_net echo-test-client nc server 12345`

if [ "$result" = "Test" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi

rm Dockerfile.Echotest
