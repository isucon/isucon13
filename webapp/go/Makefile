BUILD=go build

DOCKER_BUILD=sudo docker build
DOCKER_BUILD_OPTS=--no-cache

DOCKER_RMI=sudo docker rmi -f

DESTDIR=.
TAG=isupipe:latest

.PHONY: build
build:
	$(BUILD) -o $(DESTDIR)/isupipe -ldflags "-s -w"

.PHONY: docker_image
docker_image: clean build
	$(DOCKER_BUILD) -t $(TAG) . $(DOCKER_BUILD_OPTS)

.PHONY: clean
clean:
	$(DOCKER_RMI) -f $(TAG)
