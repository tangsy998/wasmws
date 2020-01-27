WASM_HOME=.
WASM_FILES=$(shell go env GOROOT)/misc/wasm
USER_GOPATH=/Users/fanhongling/Downloads/workspace
PROTOC_PLUGIN_DIR=${HOME}/go/bin

protobuf:
	@PATH=$(PROTOC_PLUGIN_DIR):$$PATH protoc -I/usr/local/include -I. \
		-Ivendor/github.com/grpc-ecosystem/grpc-gateway/ \
		-Ivendor/github.com/gogo/googleapis/ \
		-Ivendor/ \
		-I${USER_GOPATH}/src \
		-I${HOME}/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--gogo_out=\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types,\
Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,\
plugins=grpc:. \
		demo_server_stream.proto

build: install
	@GOOS=js GOARCH=wasm go build -o main.wasm ./demo_server_stream.pb.go ./wasm_main.go

install:
	# @cp ${WASM_FILES}/wasm_exec.js ./
	@ln -fs ${WASM_FILES}/wasm_exec.js ${WASM_HOME}/wasm_exec.js

run: build
	go run ./demo_server_stream.pb.go ./server_main.go
	
.PHONY: build install run