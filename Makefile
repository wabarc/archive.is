export GO111MODULE = on
export GOPROXY = https://proxy.golang.org

NAME = archive.is
BINDIR ?= ./build/binary
PACKDIR ?= ./build/package
LDFLAGS := $(shell echo "-X 'archive.is/version.Version=`git describe --tags --abbrev=0`'")
LDFLAGS := $(shell echo "${LDFLAGS} -X 'archive.is/version.Commit=`git rev-parse --short HEAD`'")
LDFLAGS := $(shell echo "${LDFLAGS} -X 'archive.is/version.BuildDate=`date +%FT%T%z`'")
GOBUILD ?= CGO_ENABLED=0 go build -trimpath --ldflags "-s -w ${LDFLAGS} -buildid=" -v
VERSION ?= $(shell git describe --tags `git rev-list --tags --max-count=1` | sed -e 's/v//g')
GOFILES ?= $(wildcard ./cmd/archive.is/*.go)
PROJECT := github.com/wabarc/archive.is
PACKAGES ?= $(shell go list ./...)

PLATFORM_LIST = \
	darwin-amd64 \
	linux-386 \
	linux-amd64 \
	linux-armv5 \
	linux-armv6 \
	linux-armv7 \
	linux-arm64 \
	linux-mips-softfloat \
	linux-mips-hardfloat \
	linux-mipsle-softfloat \
	linux-mipsle-hardfloat \
	linux-mips64 \
	linux-mips64le \
	linux-ppc64 \
	linux-ppc64le \
	freebsd-386 \
	freebsd-amd64 \
	openbsd-386 \
	openbsd-amd64 \
	dragonfly-amd64

WINDOWS_ARCH_LIST = \
	windows-386 \
	windows-amd64

.PHONY: \
	darwin-386 \
	darwin-amd64 \
	linux-386 \
	linux-amd64 \
	linux-armv5 \
	linux-armv6 \
	linux-armv7 \
	linux-arm64 \
	linux-mips-softfloat \
	linux-mips-hardfloat \
	linux-mipsle-softfloat \
	linux-mipsle-hardfloat \
	linux-mips64 \
	linux-mips64le \
	linux-ppc64 \
	linux-ppc64le \
	freebsd-386 \
	freebsd-amd64 \
	openbsd-386 \
	openbsd-amd64 \
	windows-386 \
	windows-amd64 \
	all-arch \
	tar_releases \
	zip_releases \
	releases \
	clean \
	test \
	fmt \
	rpm \
	debian \
	debian-packages \
	docker-image

darwin-386:
	GOARCH=386 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

darwin-arm64:
	GOARCH=arm64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-386:
	GOARCH=386 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv5:
	GOARCH=arm GOOS=linux GOARM=5 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv6:
	GOARCH=arm GOOS=linux GOARM=6 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv7:
	GOARCH=arm GOOS=linux GOARM=7 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv8: linux-arm64
linux-arm64:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips-softfloat:
	GOARCH=mips GOMIPS=softfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips-hardfloat:
	GOARCH=mips GOMIPS=hardfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mipsle-softfloat:
	GOARCH=mipsle GOMIPS=softfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mipsle-hardfloat:
	GOARCH=mipsle GOMIPS=hardfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips64:
	GOARCH=mips64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips64le:
	GOARCH=mips64le GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-ppc64:
	GOARCH=ppc64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-ppc64le:
	GOARCH=ppc64le GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

freebsd-386:
	GOARCH=386 GOOS=freebsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

freebsd-amd64:
	GOARCH=amd64 GOOS=freebsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

openbsd-386:
	GOARCH=386 GOOS=openbsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

openbsd-amd64:
	GOARCH=amd64 GOOS=openbsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

windows-386:
	GOARCH=386 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe $(GOFILES)

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe $(GOFILES)

dragonfly-amd64:
	GOARCH=amd64 GOOS=dragonfly $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

ifeq ($(TARGET),)
tar_releases := $(addsuffix .gz, $(PLATFORM_LIST))
zip_releases := $(addsuffix .zip, $(WINDOWS_ARCH_LIST))
else
ifeq ($(findstring windows,$(TARGET)),windows)
zip_releases := $(addsuffix .zip, $(TARGET))
else
tar_releases := $(addsuffix .gz, $(TARGET))
endif
endif

$(tar_releases): %.gz : %
	chmod +x $(BINDIR)/$(NAME)-$(basename $@)
	tar -czf $(PACKDIR)/$(NAME)-$(basename $@)-$(VERSION).tar.gz --transform "s/.*\///g" $(BINDIR)/$(NAME)-$(basename $@)

$(zip_releases): %.zip : %
	zip -m -j $(PACKDIR)/$(NAME)-$(basename $@)-$(VERSION).zip $(BINDIR)/$(NAME)-$(basename $@).exe

all-arch: $(PLATFORM_LIST) $(WINDOWS_ARCH_LIST)

releases: $(tar_releases) $(zip_releases)

clean:
	rm -f $(BINDIR)/*
	rm -f $(PACKDIR)/*
	rm -rf data-dir*

fmt:
	@echo "-> Running go fmt"
	@go fmt $(PACKAGES)

test:
	@echo "-> Running go test"
	@CGO_ENABLED=1 go test -v -race -cover -coverprofile=coverage.out -covermode=atomic ./...

test-integration:
	@echo 'mode: atomic' > coverage.out
	@go list ./... | xargs -n1 -I{} sh -c 'CGO_ENABLED=1 go test -race -tags=integration -covermode=atomic -coverprofile=coverage.tmp -coverpkg $(go list ./... | tr "\n" ",") {} && tail -n +2 coverage.tmp >> coverage.out || exit 255'
	@rm coverage.tmp

test-cover:
	@echo "-> Running go tool cover"
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

bench:
	@echo "-> Running benchmark"
	@go test -v -bench .

profile:
	@echo "-> Running profile"
	@go test -cpuprofile cpu.prof -memprofile mem.prof -v -bench .
