# 需要修改BINARY, 项目名字
BINARY=mydocker
GOARCH=amd64

VERSION?=?
BUILD=`date +%FT%T%z`

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

.PHONY: help linux darwin windows clean version move

help:
	@echo "usage: make <option>"
	@echo "options and effects:"
	@echo "    help   : Show help"
	@echo "    linux  : Build the linux binary of this project"
	@echo "    darwin : Build the darwin binary of this project"
	@echo "    windows: Build the windows binary of this project"
	@echo "    clean  : Remove binaries"
	@echo "    move   : Move binaries"
	@echo "    version: Display Go version"


linux:
	@GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-linux-${GOARCH} main.go

darwin:
	@GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-darwin-${GOARCH} main.go

windows:
	@GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-windows-${GOARCH}.exe main.go
clean:
	@rm -f ${BINARY}-*-{GOARCH}

move:
	@mkdir -p ./builddata && mv ${BINARY}-*-${GOARCH} builddata/ && cp -r conf/ builddata/

version:
	@go version