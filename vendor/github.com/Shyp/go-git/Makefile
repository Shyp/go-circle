.PHONY: test

install:
	go get -t -d -v ./...
	go build -v ./...

test:
	go test -v -race ./...

deps:
	godep save -r ./...
