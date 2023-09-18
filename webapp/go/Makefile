BUILD=go build

DOCKER_BUILD=sudo docker build
DOCKER_BUILD_OPTS=--no-cache

DOCKER_RMI=sudo docker rmi -f

DESTDIR=/tmp
TAG=isupipe:latest

build:
	$(BUILD) -o $(DESTDIR)/isupipe -ldflags "-s -w"
.PHONY: build

docker_image: clean build
	$(DOCKER_BUILD) -t $(TAG) . $(DOCKER_BUILD_OPTS)
.PHONY: build

clean:
	$(DOCKER_RMI) -f $(TAG)
.PHONY: clean