.PHONY: install test

STATICCHECK := $(shell command -v staticcheck)
BUMP_VERSION := $(shell command -v bump_version)

install:
	go install ./...

build:
	go get -t -d -v ./...
	go build ./...

lint:
	go vet ./...
ifndef STATICCHECK
	go get -u honnef.co/go/tools/cmd/staticcheck
endif
	staticcheck ./...

test: install lint
	go test -v -race ./...

release:
	git checkout master
ifndef BUMP_VERSION
	go get github.com/Shyp/bump_version
endif
	bump_version minor circle.go
	git push origin master
	git push origin master --tags
