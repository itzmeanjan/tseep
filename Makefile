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

docker_build_v2_server:
	docker build -t v2_server -f v2/server/Dockerfile .

docker_run_v2_server:
	docker run --name v2_server -d --env-file v2_server.env v2_server

docker_stop_v2_server:
	docker stop v2_server

docker_build_v3_server:
	docker build -t v3_server -f v3/server/Dockerfile .

docker_run_v3_server:
	docker run --name v3_server -d --env-file v3_server.env v3_server

docker_stop_v3_server:
	docker stop v3_server

docker_build_client:
	docker build -t client -f client/Dockerfile .

docker_run_client:
	docker run --name client -d --env-file client.env client

docker_stop_client:
	docker stop client
