.PHONY: install test

MEGACHECK := $(GOPATH)/bin/megacheck
BUMP_VERSION := $(GOPATH)/bin/bump_version

install:
	go install ./...

build:
	go get ./...
	go build ./...

$(MEGACHECK):
	go get -u honnef.co/go/tools/cmd/megacheck

lint: | $(MEGACHECK)
	go vet ./...
	go list ./... | grep -v vendor | xargs $(MEGACHECK)

test: lint
	go test -v -race ./...

$(BUMP_VERSION):
	go get github.com/Shyp/bump_version

release: | $(BUMP_VERSION)
	git checkout master
	$(BUMP_VERSION) minor circle.go
	git push origin master
	git push origin master --tags

equinox:
	cd circle && equinox release --version "$(shell git log -1 --pretty=%B)" --token "$(shell cat cfg/equinox)" --app app_n7HhD13kpUR --platforms darwin_amd64,linux_amd64 --signing-key ../cfg/equinox.key
