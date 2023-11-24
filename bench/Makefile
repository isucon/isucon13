BUILD=go build
TEST=go test
TEST_FLAGS=-p=1 -v

DARWIN_TARGET_ENV=GOOS=darwin GOARCH=amd64
DARWIN_TARGET_ARM64_ENV=GOOS=darwin GOARCH=arm64
LINUX_TARGET_ENV=GOOS=linux GOARCH=amd64
LINUX_TARGET_ARM64_ENV=GOOS=linux GOARCH=arm64

RM=rm -f

.PHONY: all build bench linux_amd64 test clean

all: build

build: clean
	$(DARWIN_TARGET_ENV) $(BUILD) -o ./bin/bench_darwin_amd64 ./cmd/bench
	$(DARWIN_TARGET_ARM64_ENV) $(BUILD) -o ./bin/bench_darwin_arm64 ./cmd/bench
	$(LINUX_TARGET_ENV) $(BUILD) -o ./bin/bench_linux_amd64 ./cmd/bench
	$(LINUX_TARGET_ARM64_ENV) $(BUILD) -o ./bin/bench_linux_arm64 ./cmd/bench

linux_amd64:
	$(LINUX_TARGET_ENV) $(BUILD) -o ./bin/bench_linux_amd64 ./cmd/bench

test: clean
	$(TEST) $(TEST_FLAGS) ./...

bench: build
	./bin/bench_$(shell go env GOOS)_$(shell go env GOARCH) run --dns-port=1053

pretest: build
	./bin/bench_$(shell go env GOOS)_$(shell go env GOARCH) run --dns-port=1053 --pretest-only

deploy_develop:
	sudo aws ecr get-login-password --region ap-northeast-1 | sudo docker login --username AWS --password-stdin 424484851194.dkr.ecr.ap-northeast-1.amazonaws.com
	sudo docker build -f Dockerfile.dev -t isucon13-benchmarker . --no-cache
	sudo docker tag isucon13-benchmarker:latest 424484851194.dkr.ecr.ap-northeast-1.amazonaws.com/isucon13-benchmarker:latest
	sudo docker push 424484851194.dkr.ecr.ap-northeast-1.amazonaws.com/isucon13-benchmarker:latest

deploy_production:
	sudo aws ecr get-login-password --region ap-northeast-1 | sudo docker login --username AWS --password-stdin 424484851194.dkr.ecr.ap-northeast-1.amazonaws.com
	sudo docker build -f Dockerfile.prod -t isucon13-benchmarker-prod . --no-cache
	sudo docker tag isucon13-benchmarker-prod:latest 424484851194.dkr.ecr.ap-northeast-1.amazonaws.com/isucon13-benchmarker-prod:latest
	sudo docker push 424484851194.dkr.ecr.ap-northeast-1.amazonaws.com/isucon13-benchmarker-prod:latest

clean:
	$(RM) ./bin/bench_darwin_amd64
	$(RM) ./bin/bench_darwin_arm64
	$(RM) ./bin/bench_linux_amd64
	$(RM) ./bin/bench_linux_arm64
