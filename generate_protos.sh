#!/bin/bash

set -e

mkdir -p testpb
protoc --go_out=./testpb --go_opt=module=controlplaneplayground/testpb \
       --go-grpc_out=./testpb --go-grpc_opt=module=controlplaneplayground/testpb \
       test-server/test.proto

echo "Protobuf code generated successfully." 
