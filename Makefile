DOCKER_REGISTRY := us.gcr.io/colli-304019
DOCKER_IMAGE := ${DOCKER_REGISTRY}/z4

.PHONY: proto

deploy:
	helm -n z4 \
		upgrade --install \
		--set kubePrometheusStack.enabled=true \
		z4 deployments/charts/z4

build_and_push_image: build_image push_image

push_image:
	docker push ${DOCKER_IMAGE}

build_image:
	docker build -f deployments/docker/Dockerfile -t ${DOCKER_IMAGE} .

compose_up:
	docker-compose -f deployments/docker/docker-compose.yaml up --build

compose_down:
	docker-compose -f deployments/docker/docker-compose.yaml down -v

clean_db:
	rm -rf z4data && mkdir z4data
	rm -rf z4peer && mkdir z4peer

build_and_run_m1: build_m1
	Z4_DEBUG_LOGGING_ENABLED=true \
	Z4_SERVICE_PORT=6355 \
	Z4_PEER_PORT=6356 \
	Z4_DB_DATA_DIR=./z4data \
	Z4_PROFILER_ENABLED=true \
	Z4_PEER_ID=local_peer \
	Z4_BOOTSTRAP_CLUSTER=true \
	./z4_server

build_m1: proto
	GOOS=darwin GOARCH=arm64 GO111MODULE=on go build -o z4_server cmd/server/*.go

proto:
	protoc -I proto/ \
    		--go_out proto/ \
    		--go_opt paths=source_relative \
    		--go-grpc_opt paths=source_relative \
    		--go-grpc_out proto/ \
    		proto/*.proto