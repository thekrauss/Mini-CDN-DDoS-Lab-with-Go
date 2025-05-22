#!/bin/bash

PROTO_DIR="api/proto"
OUT_DIR="."

protoc \
  -I. \
  -Ithird_party/googleapis \
  --go_out=$OUT_DIR --go_opt=paths=source_relative \
  --go-grpc_out=$OUT_DIR --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=$OUT_DIR --grpc-gateway_opt=paths=source_relative \
  $PROTO_DIR/node.proto
