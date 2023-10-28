BUILD=go build
TEST=/usr/local/go/bin/go test
TEST_FLAGS=-v
CLEANTEST=/usr/local/go/bin/go clean -testcache

DARWIN_TARGET_ENV=GOOS=darwin GOARCH=amd64
LINUX_TARGET_ENV=GOOS=linux GOARCH=amd64
LINUX_TARGET_ARM64_ENV=GOOS=linux GOARCH=arm64


DOCKER_RM=sudo docker rm -f
RM=rm -f

.PHONY: all build linux_amd64 test clean

all: build

build: clean
	$(DARWIN_TARGET_ENV) $(BUILD) -o ./bin/bench_darwin_amd64 ./cmd/bench
	$(LINUX_TARGET_ENV) $(BUILD) -o ./bin/bench_linux_amd64 ./cmd/bench
	$(LINUX_TARGET_ARM64_ENV) $(BUILD) -o ./bin/bench_linux_arm64 ./cmd/bench

linux_amd64:
	$(LINUX_TARGET_ENV) $(BUILD) -o ./bin/bench_linux_amd64 ./cmd/bench

test: clean
	sudo $(TEST) $(TEST_FLAGS) ./...

clean:
	$(RM) ./bin/bench_darwin_amd64
	$(RM) ./bin/bench_linux_amd64
	$(RM) ./bin/bench_linux_arm64
	$(DOCKER_RM) isupipe-isupipe
	$(DOCKER_RM) isupipe-scenario
	$(DOCKER_RM) mysql-isupipe
	$(DOCKER_RM) mysql-scenario
	sudo $(CLEANTEST)
