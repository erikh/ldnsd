IMAGE_NAME ?= erikh/ldnsd:testing
RELEASE_IMAGE_NAME ?= erikh/ldnsd:$(shell cat VERSION)
CODE_PATH ?= /go/src/github.com/erikh/ldnsd
GO_TEST := sudo go test -v ./... -race -count 1
GO_BENCH := sudo go test -v -bench ./... -benchtime 1m
VERSION ?= $(shell git rev-parse HEAD)

DOCKER_CMD := docker run -it \
	--rm \
	-e IN_DOCKER=1 \
	-e SETUID=$$(id -u) \
	-e SETGID=$$(id -g) \
	-w $(CODE_PATH) \
	-v ${PWD}/.go-cache:/tmp/go-build-cache \
	-v ${PWD}:$(CODE_PATH) \
	$(IMAGE_NAME)

release: distclean generate-check
	GOBIN=${PWD}/build/ldnsd-$$(cat VERSION) VERSION=$$(cat VERSION) make lint install
	# FIXME include LICENSE.md
	cp README.md example.conf build/ldnsd-$$(cat VERSION)
	cd build && tar cvzf ../ldnsd-$$(cat ../VERSION).tar.gz ldnsd-$$(cat ../VERSION)

release-image:
	VERSION=$$(cat VERSION) box -t $(RELEASE_IMAGE_NAME) box-release.rb

distclean:
	rm -rf build

install:
	GOBIN=${GOPATH}/bin go install -v github.com/golang/protobuf/protoc-gen-go
	VERSION=${VERSION} go generate -v ./...
	go install -v ./...

shell: build
	mkdir -p .go-cache
	$(DOCKER_CMD)	

build: get-box
	box -t $(IMAGE_NAME) box.rb

docker-check:
	@if [ -z "$${IN_DOCKER}" ]; then echo "You really don't want to do this"; exit 1; fi

start: docker-check stop
	sudo ldnsd example.conf &

stop: docker-check
	sudo pkill ldnsd || :

get-box:
	@if [ ! -f "$(shell which box)" ]; \
	then \
		echo "Need to install box to build the docker images we use. Requires root access."; \
		curl -sSL box-builder.sh | sudo bash; \
	fi

generate-check: generate
	git diff --exit-code

generate:
	if [ -z "$${IN_DOCKER}" ]; then make build && $(DOCKER_CMD) go generate -v ./...; else go generate -v ./...; fi

test: generate
	if [ -z "$${IN_DOCKER}" ]; then make build && $(DOCKER_CMD) $(GO_TEST); else $(GO_TEST); fi

bench: generate
	if [ -z "$${IN_DOCKER}" ]; then make build && $(DOCKER_CMD) $(GO_BENCH); else $(GO_BENCH); fi

lint:
	golangci-lint run -v

ci-protobuf:
	apt-get update -qq && apt-get install unzip curl -y
	curl -sSL -o /protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip
	unzip /protoc.zip -d /usr
	chmod -R 755 /usr/bin/protoc /usr/include/google

# these tasks account for the standard golang container image not containing protobuf tools
ci-generate: ci-protobuf
	go get -v github.com/golang/protobuf/protoc-gen-go@v1.22.0
	go generate -v ./...

ci-test: ci-generate
	go test -v -race -count 1 ./...

ci-bench: ci-generate
	go test -v -bench ./... -benchtime 1m

.PHONY: test
