GOARCH ?= "amd64"
GOOS ?= "linux"
all:
	mkdir -p dist
	cp -p sh/*.sh dist/
	GO111MODULE=off GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o dist/juun.search control/search.go
	GO111MODULE=off GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o dist/juun.fzf control/fzf.go
	GO111MODULE=off GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o dist/juun.import control/import.go
	GO111MODULE=off GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o dist/juun.service service/*.go
	GO111MODULE=off GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o dist/juun.updown control/updown.go
clean:
	rm dist/juun.* dist/*.sh
