GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
BINARY_NAME=weep
VERSION=dev-$(shell date  +%Y%m%d%H%M%S)
REGISTRY=$(REGISTRY)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

ifneq ($(BRANCH),master)
	VERSION_PRERELEASE=$(BRANCH)
endif

export VERSION
export VERSION_PRERELEASE

ifneq ($(VERSION_PRERELEASE),)
	DOCKER_TAG=v$(VERSION)-$(VERSION_PRERELEASE)
else
	DOCKER_TAG=v$(VERSION)
endif

build:
	@BINARY_NAME="$(BINARY_NAME)" sh -c "'$(CURDIR)/scripts/build.sh'"

release:
	@BINARY_NAME="$(BINARY_NAME)" LD_FLAGS='-s -w' sh -c "'$(CURDIR)/scripts/build.sh'"
	upx $(BINARY_NAME)

weep-docker:
	@BINARY_NAME="$(BINARY_NAME)-docker" GOOS=linux LD_FLAGS='-s -w -extldflags "-static"' sh -c "'$(CURDIR)/scripts/build.sh'"

build-docker: weep-docker
	docker build -t weep .
	docker tag weep:latest $(REGISTRY)/infrasec/weep:$(DOCKER_TAG)

docker: build-docker

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

fmt:
	gofmt -w $(GOFMT_FILES)

clean:
	@if [ -f $(BINARY_NAME) ] ; then rm $(BINARY_NAME) ; fi
	@if [ -f $(BINARY_NAME)-docker ] ; then rm $(BINARY_NAME)-docker ; fi
