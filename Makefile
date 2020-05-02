IMAGE_NAME ?= ldnsd:testing
CODE_PATH ?= /go/src/code.hollensbe.org/erikh/ldnsd
GO_TEST := sudo go test -v ./... -race -count 1
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

release: distclean
	GOBIN=${PWD}/build/ldnsd-$$(cat VERSION) VERSION=$$(cat VERSION) make install
	# FIXME include LICENSE.md
	cp README.md example.conf build/ldnsd-$$(cat VERSION)
	cd build && tar cvzf ../ldnsd-$$(cat ../VERSION).tar.gz ldhcpd-$$(cat ../VERSION)

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

test:
	if [ -z "$${IN_DOCKER}" ]; then make build && $(DOCKER_CMD) $(GO_TEST); else $(GO_TEST); fi

.PHONY: test
