export GOPATH=$(shell pwd)

build:
	rm -fr src/
	go get github.com/gwenn/gosqlite
