.PHONY: build release clean docker sync-consul sync-etcd sync-zk

binout=bin/aginx

Version=$(shell git describe --tags `git rev-list --tags --max-count=1`)
BuildDate=$(shell date +"%F %T")
GitCommit=$(shell git rev-parse HEAD)

debug=-w -s
param=-X main.VERSION=${Version} -X main.GITLOG_VERSION=${GitCommit} -X 'main.BUILD_TIME=${BuildDate}'

build:
	go build -ldflags "${debug} ${param}" -o ${binout}

docker:
	docker build --build-arg LDFLAGS="${debug} ${param}" -t xhaiker/aginx:${Version} -t xhaiker/aginx .

sync-consul: build
	./bin/aginx -d sync consul://127.0.0.1:8500/aginx

sync-etcd: build
	./bin/aginx -d sync etcd://127.0.0.1:2379/aginx

sync-zk: build
	./bin/aginx -d sync zk://127.0.0.1:2181/aginx

gen:
	make -C ui
	go generate ui/generator.go

release: gen
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags bindata -ldflags "${debug} ${param}" -o bin/aginx-linux-amd64-${Version} aginx.go
	tar -czvf bin/aginx-linux-amd64-${Version}.tar.gz bin/aginx-linux-amd64-${Version}

	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -tags bindata -ldflags "${debug} ${param}" -o bin/aginx-darwin-amd64-${Version} aginx.go
	tar -czvf bin/aginx-darwin-amd64-${Version}.tar.gz bin/aginx-darwin-amd64-${Version}

	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags bindata -ldflags "${debug} ${param}" -o bin/aginx-windows-amd64-${Version}.exe aginx.go
	zip bin/aginx-windows-amd64-${Version}.zip bin/aginx-windows-amd64-${Version}.exe

clean:
	@rm -rf bin

