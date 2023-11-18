export BRANCH := $(shell git branch --contains | fgrep '*' | cut -d ' ' -f 2)
.DEFAULT_GOAL := all

.PHONY: all
all:
	make app-build
	make envcheck-build
	make init
	make build

.PHONY: app-build
app-build: 
	make -C ../../webapp/go build

.PHONY: envcheck-build
envcheck-build:
	make -C ../../envcheck build

.PHONY: init
init:
	packer init -upgrade isucon13.pkr.hcl

.PHONY: build
build: app-build init
	packer build isucon13.pkr.hcl