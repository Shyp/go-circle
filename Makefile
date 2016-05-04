.PHONY: install test

install:
	go install ./...

build:
	go get ./...
	go build ./...

test: install
	go test -v -race ./...

release:
	git checkout master
	bump_version minor circle.go
	git push origin master
	git push origin master --tags
