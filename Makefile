.PHONY: fmt tidy vet test gosec go-tools pin-actions publish-check publish

LOCAL_GOCACHE := $(CURDIR)/.gocache
LOCAL_GOMODCACHE := $(CURDIR)/.gomodcache
LOCAL_GOPATH := $(CURDIR)/.gopath
LOCAL_GOBIN := $(CURDIR)/.bin
GOENV := GOCACHE=$(LOCAL_GOCACHE) GOMODCACHE=$(LOCAL_GOMODCACHE) GOPATH=$(LOCAL_GOPATH) GOBIN=$(LOCAL_GOBIN) GOTOOLCHAIN=auto
MOD_FLAG ?= -mod=readonly
PUBLISH ?= 0
RELEASE_REMOTE ?= github

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
	$(GOENV) go install github.com/suzuki-shunsuke/pinact/v4/cmd/pinact@latest

pin-actions:
	$(LOCAL_GOBIN)/pinact run --update --verify-comment

publish-check:
	@version='$(VERSION)'; \
	if [ -z "$$version" ]; then \
		echo "VERSION is required, for example: make publish VERSION=v0.1.0"; \
		exit 1; \
	fi; \
	if ! command -v git-cliff >/dev/null 2>&1; then \
		echo "git-cliff is required before publishing."; \
		exit 1; \
	fi; \
	if [ ! -x "$(LOCAL_GOBIN)/pinact" ]; then \
		echo "$(LOCAL_GOBIN)/pinact is required before publishing. Run: make go-tools"; \
		exit 1; \
	fi; \
	if ! printf '%s\n' "$$version" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$$'; then \
		echo "VERSION must be a semantic version tag such as v0.1.0 or v1.2.3-rc.1"; \
		exit 1; \
	fi; \
	if [ -n "$$(git status --porcelain)" ]; then \
		echo "Refusing to publish from a dirty worktree. Commit or stash changes first."; \
		exit 1; \
	fi; \
	if ! git remote get-url "$(RELEASE_REMOTE)" >/dev/null 2>&1; then \
		echo "The git remote '$(RELEASE_REMOTE)' is required before publishing."; \
		exit 1; \
	fi; \
	if git rev-parse "$$version" >/dev/null 2>&1; then \
		echo "Git tag $$version already exists."; \
		exit 1; \
	fi; \
	module_path="$$(awk '/^module / { print $$2; exit }' go.mod)"; \
	if [ -z "$$module_path" ]; then \
		echo "Unable to determine module path from go.mod."; \
		exit 1; \
	fi; \
	case "$$module_path" in \
		*.*) ;; \
		*) \
			echo "Module path '$$module_path' is not a publishable import path."; \
			exit 1; \
			;; \
	esac; \
	major="$${version#v}"; \
	major="$${major%%.*}"; \
	case "$$module_path" in \
		*/v[0-9]*) \
			path_major="$${module_path##*/v}"; \
			if [ "$$path_major" != "$$major" ]; then \
				echo "Module path major suffix /v$$path_major does not match VERSION $$version."; \
				exit 1; \
			fi; \
			;; \
		*) \
			if [ "$$major" -ge 2 ]; then \
				echo "VERSION $$version requires the module path to end with /v$$major."; \
				exit 1; \
			fi; \
			;; \
	esac

publish: publish-check pin-actions tidy test
	@if [ "$(PUBLISH)" != "1" ]; then \
		echo "Preflight complete. Refusing to tag or publish without PUBLISH=1."; \
		echo "When ready, run: make publish VERSION=$(VERSION) PUBLISH=1"; \
		exit 1; \
	fi
	@tag_message_file="$$(mktemp)"; \
	trap 'rm -f "$$tag_message_file"' EXIT HUP INT TERM; \
	git-cliff --unreleased --tag "$(VERSION)" > "$$tag_message_file"; \
	git tag -s --cleanup=verbatim "$(VERSION)" -F "$$tag_message_file"
	@git push $(RELEASE_REMOTE) $(VERSION)
	@module_path="$$(awk '/^module / { print $$2; exit }' go.mod)"; \
	GOPROXY=proxy.golang.org $(GOENV) go list -m "$$module_path@$(VERSION)"
