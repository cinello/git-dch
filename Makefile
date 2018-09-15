PROJECT      = git-dch
PACKAGEPATH  = gitlab.yuribugelli.it/debian

SHELL     = /bin/bash -o pipefail

_PACKAGE  = $(PACKAGEPATH)/$(PROJECT)
_DATE    ?= $(shell date "+%Y%m%d_%H%M%S_%z_%Z")
_VERSION ?= $(shell ver=$$(git tag --list  "[0-9999].[0-9999].[0-9999]" --sort=-version:refname 2>/dev/null | \
				head -n 1) ; [[ $${PIPESTATUS[0]} -eq 0 && "$${ver}" != "" ]] && echo $${ver} || \
			cat $(CURDIR)/.version 2>/dev/null || echo 0.0.0)
_HASH    ?= $(shell git rev-parse --short HEAD 2>/dev/null || \
			cat $(CURDIR)/.hash 2>/dev/null || echo -)
_GOPATH   = $(CURDIR)/.gopath~
_BIN      = $(_GOPATH)/bin
_BASE     = $(_GOPATH)/src/$(_PACKAGE)
_PKGS     = $(or $(PKG),$(shell cd $(_BASE) && env GOPATH=$(_GOPATH) $(_GO) list ./... | grep -v "^$(_PACKAGE)/vendor/"))
_TESTPKGS = $(shell env GOPATH=$(_GOPATH) $(_GO) list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(_PKGS))
_MAINPKGS = $(shell env GOPATH=$(_GOPATH) $(_GO) list -f '{{ if eq .Name "main" }}{{ .ImportPath }}{{ end }}' $(_PKGS))
_BRANCH   = $(shell git rev-parse --abbrev-ref HEAD)

_GO      = go
_GODOC   = godoc
_GOFMT   = gofmt
_TIMEOUT = 15
_UPX     = upx
_V = 0
_Q = $(if $(filter 1,$(_V)),,@)
_M = $(shell printf "=>")
_M2 = $(shell printf "  ->")

_HOSTARCH = $(shell env GOPATH=$(_GOPATH) $(_GO) env | grep GOHOSTARCH | awk -F'=' '{print $$2}')
_HOSTOS = $(shell env GOPATH=$(_GOPATH) $(_GO) env | grep GOHOSTOS | awk -F'=' '{print $$2}')

.DEFAULT_GOAL := build

$(_BASE): ; $(info $(_M) setting GOPATH...)
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR) $@

# Tools

_MEGACHECK = $(_BIN)/megacheck
$(_BIN)/megacheck: | $(_BASE) ; $(info $(_M) building megacheck...)
	$(_Q) GOPATH=$(_GOPATH) GOOS=$(_HOSTOS) GOARCH=$(_HOSTARCH) $(_GO) get honnef.co/go/tools/cmd/megacheck

_GOCOVMERGE = $(_BIN)/gocovmerge
$(_BIN)/gocovmerge: | $(_BASE) ; $(info $(_M) building gocovmerge...)
	$(_Q) GOPATH=$(_GOPATH) GOOS=$(_HOSTOS) GOARCH=$(_HOSTARCH) $(_GO) get github.com/wadey/gocovmerge

_GOCOV = $(_BIN)/gocov
$(_BIN)/gocov: | $(_BASE) ; $(info $(_M) building gocov...)
	$(_Q) GOPATH=$(_GOPATH) GOOS=$(_HOSTOS) GOARCH=$(_HOSTARCH) $(_GO) get github.com/axw/gocov/...

_GOCOVXML = $(_BIN)/gocov-xml
$(_BIN)/gocov-xml: | $(_BASE) ; $(info $(_M) building gocov-xml...)
	$(_Q) GOPATH=$(_GOPATH) GOOS=$(_HOSTOS) GOARCH=$(_HOSTARCH) $(_GO) get github.com/AlekSi/gocov-xml

_GO2XUNIT = $(_BIN)/go2xunit
$(_BIN)/go2xunit: | $(_BASE) ; $(info $(_M) building go2xunit...)
	$(_Q) GOPATH=$(_GOPATH) GOOS=$(_HOSTOS) GOARCH=$(_HOSTARCH) $(_GO) get github.com/tebeka/go2xunit

_DEP = $(_BIN)/dep
$(_BIN)/dep: | $(_BASE) ; $(info $(_M) building dep...)
	$(_Q) GOPATH=$(_GOPATH) GOOS=$(_HOSTOS) GOARCH=$(_HOSTARCH) $(_GO) get github.com/golang/dep/cmd/dep

# Tests

_TEST_TARGETS := test-default test-bench test-short test-verbose test-race
.PHONY: $(_TEST_TARGETS) test-xml check test tests
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
$(_TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(_TEST_TARGETS): test
check test tests: export GOPATH=$(_GOPATH)
check test tests: vendor | $(_BASE) ; $(info $(_M) running $(NAME:%=% )tests...) @ ## Run tests
	$(_Q) cd $(_BASE) && if [ "$(_TESTPKGS)" != "" ]; then $(_GO) test -timeout $(_TIMEOUT)s $(ARGS) $(_TESTPKGS); fi

test-xml: export GOPATH=$(_GOPATH)
test-xml: vendor | $(_BASE) $(_GO2XUNIT) ; $(info $(_M) running $(NAME:%=% )tests...) @ ## Run tests with xUnit output
	$(_Q) if [ ! -d test ]; then mkdir test; fi
	$(_Q) if [ "$(_TESTPKGS)" != "" ]; then \
		cd $(_BASE) && 2>&1 $(_GO) test -timeout 20s -v $(_TESTPKGS) | tee test/tests.output && \
		$(_GO2XUNIT) -fail -input test/tests.output -output test/tests.xml; \
	 fi

_COVERAGE_MODE = atomic
_COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
_COVERAGE_XML = $(COVERAGE_DIR)/coverage.xml
_COVERAGE_HTML = $(COVERAGE_DIR)/index.html
.PHONY: test-coverage test-coverage-tools
test-coverage-tools: | $(_GOCOVMERGE) $(_GOCOV) $(_GOCOVXML)
test-coverage: export GOPATH=$(_GOPATH)
test-coverage: COVERAGE_DIR := $(CURDIR)/test/coverage.$(shell date -u +"%Y-%m-%dT%H.%M.%SZ")
test-coverage: vendor test-coverage-tools | $(_BASE) ; $(info $(_M) running coverage tests...) @ ## Run coverage tests
	$(_Q) mkdir -p $(COVERAGE_DIR)/coverage
	$(_Q) cd $(_BASE) && for pkg in $(_TESTPKGS); do \
		$(_GO) test \
			-coverpkg=$$($(_GO) list -f '{{ join .Deps "\n" }}' $$pkg | \
					grep '^$(_PACKAGE)/' | grep -v '^$(_PACKAGE)/vendor/' | \
					tr '\n' ',')$$pkg \
			-covermode=$(_COVERAGE_MODE) \
			-coverprofile="$(COVERAGE_DIR)/coverage/`echo $$pkg | tr "/" "-"`.cover" $$pkg ;\
	 done
	$(_Q) $(_GOCOVMERGE) $(COVERAGE_DIR)/coverage/*.cover > $(_COVERAGE_PROFILE)
	$(_Q) $(_GO) tool cover -html=$(_COVERAGE_PROFILE) -o $(_COVERAGE_HTML)
	$(_Q) $(_GOCOV) convert $(_COVERAGE_PROFILE) | $(_GOCOVXML) > $(_COVERAGE_XML)

.PHONY: megacheck
megacheck: export GOPATH=$(_GOPATH)
megacheck: vendor | $(_BASE) $(_MEGACHECK) ; $(info $(_M) running megacheck...) @ ## Run megacheck
	$(_Q) cd $(_BASE) && ret=0 && for pkg in $(_PKGS); do \
		test -z "$$($(_MEGACHECK) $$pkg | tee /dev/stderr)" || ret=1 ; \
	 done ; exit $$ret

.PHONY: fmt
fmt: export GOPATH=$(_GOPATH)
fmt: ; $(info $(_M) running gofmt...) @ ## Run gofmt on all source files
	@ret=0 && for d in $$(GOPATH=$(_GOPATH) $(_GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(_GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

# Dependency management

dep.lock: Gopkg.toml | $(_BASE) ; $(info $(_M) updating dependencies...)
	$(_Q) cd $(_BASE) && GOPATH=$(_GOPATH) $(_DEP) ensure -update
	touch $@
vendor: Gopkg.lock | $(_BASE) $(_DEP) ; $(info $(_M) retrieving dependencies...)
	$(_Q) cd $(_BASE) && GOPATH=$(_GOPATH) $(_DEP) ensure
	@touch $@

# Misc

.PHONY: clean
clean: ; $(info $(_M) cleaning binaries...)	@ ## Cleanup
	@rm -rf bin
	@rm -rf test

.PHONY: clean-all
clean-all: clean ; $(info $(_M) cleaning dependencies...)	@ ## Cleanup everything
	@rm -rf $(_GOPATH)
	@rm -rf $(_BUILDDIR)
	@rm -rf vendor

.PHONY: help
help:
	@echo "TARGETS"
	@echo "---------"
	@grep -E '^[ 0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "OPTIONS"
	@echo "---------"
	@grep -E '^[ 0-9A-Z_-]+=.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = "=.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version: ## Print project version
	$(_Q) echo $(PROJECT) version $(_VERSION)

########################### BUILS #############################

GOOS        =## Cross-compile target OS
GOARCH      =## Cross-compile target architecture
_DEFLDFLAGS = -w -s -X main.version=$(_VERSION) -X main.buildDate=$(_DATE) -X main.commitHash=$(_HASH)
LDFLAGS     =## Cross-compile external linking LDFLAGS
COMPRESS    =## Set to true to compress executables

_BUILD_TARGET_DIR = $(shell export GOOS=$(GOOS) && export GOARCH=$(GOARCH) && echo $$($(_GO) env GOOS)_$$($(_GO) env GOARCH))
_BUILD_TARGET_EXE = $(shell export GOOS=$(GOOS) && export GOARCH=$(GOARCH) && echo $$($(_GO) env GOEXE))

.PHONY: build
build: export GOPATH=$(_GOPATH)
build: vendor | $(_BASE) ; $(info $(_M) building executables...) @ ## Build all the executables in this project
	$(_Q) cd $(_BASE)
	$(_Q) for pkg in $(_MAINPKGS); do \
		output=$$(basename $$pkg) && \
		echo "  compiling $(CURDIR)/bin/$(_BUILD_TARGET_DIR)/$${output}$(_BUILD_TARGET_EXE)" && \
		CGO_ENABLED=0 $(_GO) build \
			-o $(CURDIR)/bin/$(_BUILD_TARGET_DIR)/$${output}$(_BUILD_TARGET_EXE) \
			-tags release \
			-ldflags '-extldflags "-static" $(_DEFLDFLAGS) $(LDFLAGS)' \
			$${pkg} && \
		if [ "$(COMPRESS)" != "" ]; then \
			echo "  compressing $(CURDIR)/bin/$(_BUILD_TARGET_DIR)/$${output}$(_BUILD_TARGET_EXE)" && \
			$(_UPX) -9 -qq $(CURDIR)/bin/$(_BUILD_TARGET_DIR)/$${output}$(_BUILD_TARGET_EXE); \
	 	fi; \
	 done

.PHONY: env
env: | $(_BASE) ; $(info $(_M) printing go environment...) ## Print go environment variables
	$(_Q) GOPATH=$(_GOPATH) $(_GO) env
