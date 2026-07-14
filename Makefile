.PHONY: fmt tidy vet test gosec go-tools

LOCAL_GOCACHE := $(CURDIR)/.gocache
LOCAL_GOMODCACHE := $(CURDIR)/.gomodcache
LOCAL_GOPATH := $(CURDIR)/.gopath
LOCAL_GOBIN := $(CURDIR)/.bin
GOENV := GOCACHE=$(LOCAL_GOCACHE) GOMODCACHE=$(LOCAL_GOMODCACHE) GOPATH=$(LOCAL_GOPATH) GOBIN=$(LOCAL_GOBIN) GOTOOLCHAIN=auto
MOD_FLAG ?= -mod=readonly

fmt:
	@mkdir -p $(LOCAL_GOCACHE) $(LOCAL_GOMODCACHE) $(LOCAL_GOPATH) $(LOCAL_GOBIN)
	$(GOENV) go fmt ./...

tidy:
	@mkdir -p $(LOCAL_GOCACHE) $(LOCAL_GOMODCACHE) $(LOCAL_GOPATH) $(LOCAL_GOBIN)
	$(GOENV) go mod tidy

vet:
	@mkdir -p $(LOCAL_GOCACHE) $(LOCAL_GOMODCACHE) $(LOCAL_GOPATH) $(LOCAL_GOBIN)
	$(GOENV) go vet $(MOD_FLAG) ./...

test: fmt vet
	@mkdir -p $(LOCAL_GOCACHE) $(LOCAL_GOMODCACHE) $(LOCAL_GOPATH) $(LOCAL_GOBIN)
	$(GOENV) go mod verify
	$(GOENV) go test $(MOD_FLAG) ./...

gosec:
	$(GOENV) $(LOCAL_GOBIN)/gosec -exclude-dir=.bin -exclude-dir=.gocache -exclude-dir=.gomodcache -exclude-dir=.gopath -exclude-dir=.git ./...
	$(GOENV) $(LOCAL_GOBIN)/govulncheck -scan module

go-tools:
	@mkdir -p $(LOCAL_GOCACHE) $(LOCAL_GOMODCACHE) $(LOCAL_GOPATH) $(LOCAL_GOBIN)
	$(GOENV) go install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GOENV) go install golang.org/x/vuln/cmd/govulncheck@latest
