.PHONY: build release clean

binout=bin/aginx

ifeq ($(P),release)
Version=$(shell git describe --tags `git rev-list --tags --max-count=1`)
BuildDate=$(shell date +"%F %T")
GitCommit=$(shell git rev-parse HEAD)
debug=-w -s
param=-X main.VERSION=${Version} -X main.GITLOG_VERSION=${GitCommit} -X 'main.BUILD_TIME=${BuildDate}'
else
debug=
param=
endif

build:
	go mod download
	go build -tags bindata -ldflags "${debug} ${param}" -o ${binout}

release:
	make -C . -e P=release

clean:
	@rm -rf bin

