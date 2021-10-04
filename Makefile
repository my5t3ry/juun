GOARCH ?= "amd64"
GOOS ?= "linux"
all:
	mkdir -p dist
	cp -p sh/*.sh dist/
	GO111MODULE=off GOARCH=$(GOARCH) GOOS=$(GOOS) go build -ldflags="-s -w" -o dist/juun.search control/search.go
	GO111MODULE=off GOARCH=$(GOARCH) GOOS=$(GOOS) go build -ldflags="-s -w" -o dist/juun.complete control/complete.go
	GO111MODULE=off GOARCH=$(GOARCH) GOOS=$(GOOS) go build -ldflags="-s -w" -o dist/juun.import control/import.go
	GO111MODULE=off GOARCH=$(GOARCH) GOOS=$(GOOS) go build -ldflags="-s -w" -o dist/juun.service service/*.go
	GO111MODULE=off GOARCH=$(GOARCH) GOOS=$(GOOS) go build -ldflags="-s -w" -o dist/juun.updown control/updown.go
clean:
	rm dist/juun.* dist/*.sh
