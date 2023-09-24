BUILD=go build
TEST=/usr/local/go/bin/go test
TEST_FLAGS=-v -p=1

DARWIN_TARGET_ENV=GOOS=darwin GOARCH=amd64
LINUX_TARGET_ENV=GOOS=linux GOARCH=amd64


DOCKER_RM=sudo docker rm -f
RM=rm -f

.PHONY: all build test clean

all: build

build: clean
	$(DARWIN_TARGET_ENV) $(BUILD) -o ./bin/bench_darwin ./cmd/bench
	$(LINUX_TARGET_ENV) $(BUILD) -o ./bin/bench_linux ./cmd/bench

test: clean
	sudo $(TEST) $(TEST_FLAGS) ./...

clean:
	$(RM) ./bin/bench_darwin
	$(RM) ./bin/bench_linux
	$(DOCKER_RM) isupipe
	$(DOCKER_RM) mysql
