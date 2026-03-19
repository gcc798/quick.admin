KRATOS_ROOT := $(CURDIR)
THIRD_PARTY_DIR := $(KRATOS_ROOT)/third_party
SYS_API_DIR := $(KRATOS_ROOT)/application/sys-api
SYS_RPC_DIR := $(KRATOS_ROOT)/application/sys-rpc
SYS_API_SERVICE_GEN_DIR := $(SYS_API_DIR)/internal/servicegen
SYS_RPC_SERVICE_GEN_DIR := $(SYS_RPC_DIR)/internal/servicegen
PROTO_FILES := $(shell cd $(KRATOS_ROOT) && find api -name '*.proto' | sort)
PROTOBUF_INCLUDE := /opt/homebrew/include/google/protobuf

export PATH := /Users/guoc/dev/code_go/bin:$(PATH)
export GOCACHE := /tmp/nai-tizi-kratos-go-build

.PHONY: help init prepare-third-party clean-proto proto openapi proto-all proto-add-health proto-client proto-server-sys-api proto-server-sys-rpc ent wire fmt build-sys-api build-sys-rpc build-all test

help:
	@echo "targets: init prepare-third-party clean-proto proto openapi proto-all proto-add-health proto-client proto-server-sys-api proto-server-sys-rpc ent wire fmt build-sys-api build-sys-rpc build-all test"

init: prepare-third-party
	go mod tidy

prepare-third-party:
	mkdir -p $(THIRD_PARTY_DIR)/google/api $(THIRD_PARTY_DIR)/google/protobuf $(THIRD_PARTY_DIR)/validate
	cp /Users/guoc/dev/code_go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/google/api/annotations.proto $(THIRD_PARTY_DIR)/google/api/
	cp /Users/guoc/dev/code_go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/google/api/http.proto $(THIRD_PARTY_DIR)/google/api/
	cp $(PROTOBUF_INCLUDE)/descriptor.proto $(THIRD_PARTY_DIR)/google/protobuf/
	cp $(PROTOBUF_INCLUDE)/duration.proto $(THIRD_PARTY_DIR)/google/protobuf/
	cp $(PROTOBUF_INCLUDE)/struct.proto $(THIRD_PARTY_DIR)/google/protobuf/
	cp $(PROTOBUF_INCLUDE)/timestamp.proto $(THIRD_PARTY_DIR)/google/protobuf/
	cp /Users/guoc/dev/code_go/pkg/mod/github.com/envoyproxy/protoc-gen-validate@v1.0.4/validate/validate.proto $(THIRD_PARTY_DIR)/validate/

clean-proto:
	rm -rf $(KRATOS_ROOT)/api/system/v1/*.pb.go $(KRATOS_ROOT)/api/system/v1/*_grpc.pb.go $(KRATOS_ROOT)/api/system/v1/*_http.pb.go $(KRATOS_ROOT)/api/system/v1/*_pb.validate.go $(KRATOS_ROOT)/api/system/v1/*_client.go $(KRATOS_ROOT)/system $(KRATOS_ROOT)/api/openapi.yaml

proto-add-health:
	kratos proto add api/system/v1/health.proto

proto:
	protoc \
		-I . \
		-I $(THIRD_PARTY_DIR) \
		--go_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:. \
		--go-http_out=paths=source_relative:. \
		--go-errors_out=paths=source_relative:. \
		--validate_out=lang=go,paths=source_relative:. \
		$(PROTO_FILES)

openapi:
	protoc \
		-I . \
		-I $(THIRD_PARTY_DIR) \
		--openapi_out=fq_schema_naming=true,default_response=false:$(KRATOS_ROOT)/api \
		$(PROTO_FILES)

proto-all: clean-proto prepare-third-party proto openapi proto-client proto-server-sys-api proto-server-sys-rpc

proto-client:
	for file in $(PROTO_FILES); do \
		kratos proto client $$file -p $(THIRD_PARTY_DIR); \
	done

proto-server-sys-api:
	for file in $(PROTO_FILES); do \
		mkdir -p $(SYS_API_SERVICE_GEN_DIR); \
		kratos proto server $$file -t $(SYS_API_SERVICE_GEN_DIR); \
	done

proto-server-sys-rpc:
	for file in $(PROTO_FILES); do \
		mkdir -p $(SYS_RPC_SERVICE_GEN_DIR); \
		kratos proto server $$file -t $(SYS_RPC_SERVICE_GEN_DIR); \
	done

ent:
	ent generate ./application/sys-rpc/ent/schema

wire:
	GOSUMDB=off go run github.com/google/wire/cmd/wire ./application/sys-api/cmd/server ./application/sys-rpc/cmd/server

fmt:
	go fmt ./...

build-sys-api:
	go build -o $(KRATOS_ROOT)/bin/sys-api ./application/sys-api/cmd/server

build-sys-rpc:
	go build -o $(KRATOS_ROOT)/bin/sys-rpc ./application/sys-rpc/cmd/server

build-all: build-sys-api build-sys-rpc

test:
	go test ./...
