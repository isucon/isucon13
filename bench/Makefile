BUILD=go build
TEST=/usr/local/go/bin/go test
TEST_FLAGS=-v

DARWIN_TARGET_ENV=GOOS=darwin GOARCH=amd64
LINUX_TARGET_ENV=GOOS=linux GOARCH=amd64


RM=rm -f

.PHONY: all build test clean

all: build

build: clean
	$(DARWIN_TARGET_ENV) $(BUILD) -o ./bin/bench_darwin ./cmd/bench
	$(LINUX_TARGET_ENV) $(BUILD) -o ./bin/bench_linux ./cmd/bench

test:
	sudo $(TEST) $(TEST_FLAGS) ./...

clean:
	$(RM) ./bin/bench_darwin
	$(RM) ./bin/bench_linux
