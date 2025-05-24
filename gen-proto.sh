#!/bin/bash

set -e

PROTO_DIR=shared-proto
GOOGLEAPIS_DIR=third_party/googleapis

# Génération pour control-plane
protoc -I${PROTO_DIR} -I${GOOGLEAPIS_DIR} \
  --go_out=control-plane/proto --go_opt=paths=source_relative \
  --go-grpc_out=control-plane/proto --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=control-plane/proto --grpc-gateway_opt=paths=source_relative \
  ${PROTO_DIR}/node.proto

# Génération pour worker-node
protoc -I${PROTO_DIR} -I${GOOGLEAPIS_DIR} \
  --go_out=worker-node/proto --go_opt=paths=source_relative \
  --go-grpc_out=worker-node/proto --go-grpc_opt=paths=source_relative \
  ${PROTO_DIR}/node.proto

echo "✅ Protobufs générés avec succès"
