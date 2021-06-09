SHELL := /bin/bash

build_v1_server:
	pushd v1/server; go build -o ../../v1_server; popd

run_v1_server: build_v1_server
	./v1_server

build_v1_client:
	pushd v1/client; go build -o ../../v1_client; popd

run_v1_client: build_v1_client
	./v1_client

docker_build_v1_server:
	docker build -t v1_server -f v1/server/Dockerfile .

docker_run_v1_server:
	docker run --name v1_server -d --env-file v1_server.env v1_server

docker_stop_v1_server:
	docker stop v1_server
