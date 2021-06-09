SHELL := /bin/bash

build_v1:
	pushd v1/server; go build -o ../../v1_server; popd

run_v1: build_v1
	./v1_server
