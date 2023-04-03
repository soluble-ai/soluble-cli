SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

default: help

GOLANGCILINTVERSION?=1.50.0

SHELL=/bin/bash
VERSION?=$(shell git describe --tags --dirty --always)
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%S+00:00)
LDFLAGS=-X 'github.com/soluble-ai/soluble-cli/pkg/version.Version=${VERSION}' -X 'github.com/soluble-ai/soluble-cli/pkg/version.BuildTime=${BUILD_TIME}'

.PHONY: help
help:
	@echo "-------------------------------------------------------------------"
	@echo "Soluble-cli Makefile helper:"
	@echo ""
	@echo version = $(VERSION)
	@echo ""
	@grep -Fh "##" $(MAKEFILE_LIST) | grep -v grep | sed -e 's/\\$$//' | sed -E 's/^([^:]*):.*##(.*)/ \1 -\2/'
	@echo "-------------------------------------------------------------------"

none: help


.PHONY: version
version: ## Print current version
	@echo "Build version: $(VERSION)"

.PHONY: prepare
prepare: install-tools go-mod go-gen version source-name-verify

.PHONY: go-mod
go-mod: ## Runs go mod tidy, vendor and verify to cleanup, copy and verify dependencies
	go mod tidy -v
	go mod verify

.PHONY: go-gen
go-gen: ## go generate all files
	go generate ./...

.PHONY: integration-test-verify
integration-test-verify: ## verify that integeration tests sources have the correct build constraints
	@if find . -name '*.go' | egrep "integration/.*_test.go" | xargs egrep -c "//go:build integration" | egrep ":0"; then \
		echo "Error: the integration tests listed above should have a '//go:build integration' build constraint"; \
		exit 1; \
	fi

.PHONY: source-name-verify
source-name-verify: ## verify that go source files don't have - in them
	@if find . -name '*.go' | xargs -n 1 basename | egrep -e -; then \
    	echo "Error: The go source files listed above should use _ rather than - in their names"; \
    	exit 1; \
	fi

.PHONY: coverage
coverage: prepare ## go unit tests
	go test -cover ./...

.PHONY: lint
lint: prepare ## lint all code
	golangci-lint run --timeout 120s -E stylecheck -E gosec -E goimports -E misspell -E gocritic -E whitespace -E goprintffuncname -e G402;

.PHONY: integration-test-configure
integration-test-configure: ## configure integration test for github action, do not use in local dev
	@echo "Configuring an IAC profile for integration testing";
	go run main.go configure set-profile --quiet --format none integ-test
	go run main.go configure set --quiet --format none APIToken ${SOLUBLE_API_TOKEN}
	go run main.go configure set --quiet --format none APIServer ${SOLUBLE_API_SERVER}
	go run main.go configure set --quiet --format none Organization ${SOLUBLE_ORGANIZATION}
	go run main.go configure show

.PHONY: integration-test
integration-test: integration-test-verify coverage ## run integration test suite
	@if go run main.go configure show --format 'value(ProfileName)' | egrep -e '-test' > /dev/null; then \
		echo "Running go test (integration tests)"; \
		go test -tags=integration -timeout 60s ./.../integration; \
	else \
		echo "Skipping integration tests because the current profile does not end in -test"; \
		exit 1; \
	fi \

.PHONY: dist
dist: ## build binary with optional file extension (ext) and package (pkg) for given os and arch
	make bin os=$(os) arch=$(arch) ext=$(ext)
	mkdir -p dist
	cp LICENSE README.md target/$(os)_$(arch)
	$(eval name=soluble_$(VERSION)_$(os)_$(arch))
	@echo "Packaging $(name)"
	@if [ "$(pkg)" = "tar" ]; then \
		tar cvf ./dist/$(name).tar.gz --use-compress-program='gzip -9' -C target/$(os)_$(arch) .; \
	elif [ "$(pkg)" = "zip" ]; then \
		zip -j ./dist/$(name).zip target/$(os)_$(arch)/*; \
	fi

.PHONY: bin
bin: ## build binary with optional file extension (ext) for a specific os and arch
	@echo "Building soluble binary for $(os) $(arch)"
	rm -rf target/$(os)_$(arch)
	mkdir -p target/$(os)_$(arch)
	GOOS=$(os) GOARCH=$(arch) go build -o target/$(os)_$(arch)/soluble$(ext) -tags ci,osusergo,netgo -trimpath -ldflags="$(LDFLAGS) $(ldflags)"

.PHONY: dist-clean
dist-clean:
	rm -rf ./dist

.PHONY: dist-all ## build all binaries and packages
dist-all: dist-clean \
	lint \
	linux-amd64-tar \
	linux-arm64-tar \
	darwin-amd64-tar \
	darwin-arm64-tar \
	windows-amd64-zip

.PHONY: dist-all-test
dist-all-test: integration-test dist-all ## run all tests, build all binaries and packages

.PHONY: install-darwin-iac
install-darwin-iac: darwin-amd64-tar ## Convenience target to build and deploy the lacework iac component for local MAC development
	make bin os="darwin" arch="amd64"
	mkdir -p ~/.config/lacework/components/iac
	VERSION=$(VERSION) envsubst < scripts/.dev.template > $$HOME/.config/lacework/components/iac/.dev
	cp target/darwin_amd64/soluble $$HOME/.config/lacework/components/iac/iac

.PHONY: uninstall-darwin-iac
uninstall-darwin-iac:## Uninstall the lacework iac component
	@rm -rf ~/.config/lacework/components/iac
	@echo "Development version uninstalled, run:"
	@echo
	@echo "  lacework components install iac"
	@echo
	@echo "to restore the old version"

.PHONY: linux-amd64-tar
linux-amd64-tar:
	make dist os="linux" arch="amd64" pkg="tar" ldflags="-extldflags -static"

.PHONY: linux-arm64-tar
linux-arm64-tar:
	make dist os="linux" arch="arm64" pkg="tar" ldflags="-extldflags -static"

.PHONY: darwin-arm64-tar
darwin-arm64-tar:
	make dist os="darwin" arch="arm64" pkg="tar" ext="" ldflags=""

.PHONY: darwin-amd64-tar
darwin-amd64-tar:
	make dist os="darwin" arch="amd64" pkg="tar"  ext="" ldflags=""

.PHONY: windows-amd64-zip
windows-amd64-zip:
	make dist os="windows" arch="amd64" pkg="zip" ext=".exe" ldflags=""

.PHONY: install-tools
install-tools: ## Install go indirect dependencies
ifeq (, $(shell which golangci-lint))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v$(GOLANGCILINTVERSION)
endif

.PHONY: uninstall-tools
uninstall-tools: ## Uninstall go indirect dependencies
ifneq (, $(shell which golangci-lint))
	rm $(shell go env GOPATH)/bin/golangci-lint
endif