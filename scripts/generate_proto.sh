#!/bin/bash

echo "Generating protobuf files..."

# Navigate to project root
cd "$(dirname "$0")/.."

# Generate Go protobuf files
echo "Generating Go protobuf files..."
mkdir -p pkg/pb
protoc -Iproto --go_out=pkg/pb --go_opt=paths=source_relative --go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative proto/user.proto

# Generate Python protobuf files
echo "Generating Python protobuf files..."
cd client
uv run python -m grpc_tools.protoc -I../proto --python_out=proto --grpc_python_out=proto ../proto/user.proto

echo "Protobuf files generated successfully!"
echo "Go files: pkg/pb/"
echo "Python files: client/proto/"