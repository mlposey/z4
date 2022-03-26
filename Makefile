.PHONY: proto

build_and_run_m1: build_m1
	DEBUG_LOGGING_ENABLED=1 ./z4_server

build_m1: proto
	GOOS=darwin GOARCH=arm64 GO111MODULE=on go build -o z4_server cmd/server/main.go

proto:
	protoc -I proto/ \
    		--go_out proto/ \
    		--go_opt paths=source_relative \
    		--go-grpc_opt paths=source_relative \
    		--go-grpc_out proto/ \
    		proto/*.proto