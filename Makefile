TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
TAG := $(shell git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
VERSION := $(TAG:v%=%)
ifneq ($(COMMIT), $(TAG_COMMIT))
    VERSION := $(VERSION)-next-$(COMMIT)-$(DATE)
endif
ifeq ($(VERSION), )
    VERSION := $(COMMIT)-$(DATA)
endif
ifneq ($(shell git status --porcelain),)
    VERSION := $(VERSION)-dirty
endif
LDFLAGS := -ldflags "-w -s -X 'github.com/vulcanize/gap-filler-service/cmd.version=$(VERSION)'"

all: clean test linux darwin windows

clean:
	if [ -d "build" ]; then rm -rf "build"; fi

test:
	go vet ./...
	go fmt ./...
	go test ./...

linux:
	if [ ! -d "build" ]; then mkdir "build"; fi
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o build/gapfiller-linux-${COMMIT}

darwin:
	if [ ! -d "build" ]; then mkdir "build"; fi
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o build/gapfiller-darwin-${COMMIT}

windows:
	if [ ! -d "build" ]; then mkdir "build"; fi
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o build/gapfiller-windows-${COMMIT}

## Build docker image
.PHONY: docker-build
docker-build:
	docker build -t vulcanize/gap-filler -f Dockerfile .