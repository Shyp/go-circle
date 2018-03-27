.PHONY: install build test
BUMP_VERSION := $(GOPATH)/bin/bump_version
GODOCDOC := $(GOPATH)/bin/godocdoc
MEGACHECK := $(GOPATH)/bin/megacheck
UNAME = $(shell uname -s)

install:
	go get ./...
	go install ./...

$(MEGACHECK):
ifeq ($(UNAME),Darwin)
	curl --silent --location --output $(MEGACHECK) https://github.com/kevinburke/go-tools/releases/download/2018-01-25/megacheck-darwin-amd64
else
	curl --silent --location --output $(MEGACHECK) https://github.com/kevinburke/go-tools/releases/download/2018-01-25/megacheck-linux-amd64
endif
	chmod +x $(MEGACHECK)


lint: $(MEGACHECK)
	$(MEGACHECK) ./...
	go vet ./...

test: lint
	go test ./...

race-test:
	go test -race ./...

$(BUMP_VERSION):
	go get -u github.com/Shyp/bump_version

release: race-test | $(BUMP_VERSION)
	$(BUMP_VERSION) minor types.go

$(GODOCDOC):
	go get -u github.com/kevinburke/godocdoc

docs: | $(GODOCDOC)
	$(GODOCDOC)
