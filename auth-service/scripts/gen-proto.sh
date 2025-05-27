#!/bin/bash

set -e

protoc \
  --go_out=proto/ \
  --go-grpc_out=proto/ \
  --grpc-gateway_out=proto/ \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_opt=paths=source_relative \
  -I proto/ \
  -I proto/third_party/googleapis \
  proto/*.proto


