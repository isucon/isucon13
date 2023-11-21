DARWIN_TARGET_ENV=GOOS=darwin GOARCH=arm64
LINUX_TARGET_ENV=GOOS=linux GOARCH=amd64

BUILD=go build

DESTDIR=.
TAG=envcheck:latest

.PHONY: build
build:
	CGO_ENABLED=0 $(DARWIN_TARGET_ENV) $(BUILD) -o $(DESTDIR)/envcheck_darwin -ldflags "-s -w"
	CGO_ENABLED=0 $(LINUX_TARGET_ENV)  $(BUILD) -o $(DESTDIR)/envcheck -ldflags "-s -w"
	cp -a envcheck ../provisioning/ansible/roles/envcheck/files/
