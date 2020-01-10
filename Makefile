.SHELL := /bin/bash
.DEFAULT_GOAL := build
.PHONY = all tests clean

clean:
	rm -f authenticating-route-service

tests:
	go vet -v
	./scripts/test.sh

deploy: tests
	scripts/deploy.sh

build: clean tests
	go build
