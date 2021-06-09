SHELL := /bin/bash

build_v1_server:
	pushd v1/server; go build -o ../../v1_server; popd

run_v1_server: build_v1_server
	./v1_server

build_v1_client:
	pushd v1/client; go build -o ../../v1_client; popd

run_v1_client: build_v1_client
	./v1_client
