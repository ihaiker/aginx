.PHONY: build release clean docker gen

#Version=$(shell git describe --tags `git rev-list --tags --max-count=1`)
Version="v2.0"
#GitCommit=$(shell git rev-parse HEAD)
GitCommit="0000"
BuildDate=$(shell date +"%F %T")

debug=-w -s
param=-X main.VERSION=${Version} -X main.GITLOG_VERSION=${GitCommit} -X 'main.BUILD_TIME=${BuildDate}'

binout=../bin/aginx-${Version}

build:
	go build -tags bindata -ldflags "${debug} ${param}" -o ${binout}

docker:
	docker build --build-arg LDFLAGS="${debug} ${param}" -t xhaiker/aginx:${Version} -t xhaiker/aginx .

gen:
	go generate core/admin/generator_static.go

release:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags bindata -ldflags "${debug} ${param}" -o ${binout}-linux-amd64 aginx.go
	tar -czvf ${binout}-linux-amd64.tar.gz ${binout}-linux-amd64

	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -tags bindata -ldflags "${debug} ${param}" -o ${binout}-darwin-amd64 aginx.go
	tar -czvf ${binout}-darwin-amd64.tar.gz ${binout}-darwin-amd64

	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags bindata -ldflags "${debug} ${param}" -o ${binout}-windows-amd64.exe aginx.go
	zip ${binout}-windows-amd64.zip ${binout}-windows-amd64.exe

