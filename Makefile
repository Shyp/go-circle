.PHONY: install test

STATICCHECK := $(shell command -v staticcheck)
BUMP_VERSION := $(shell command -v bump_version)

install:
	go install ./...

build:
	go get ./...
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

equinox:
	cd circle && equinox release --version "$(shell git log -1 --pretty=%B)" --token "$(shell cat cfg/equinox)" --app app_n7HhD13kpUR --platforms darwin_amd64,linux_amd64 --signing-key ../cfg/equinox.key
