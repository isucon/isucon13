export BRANCH := $(shell git branch --contains | fgrep '*' | cut -d ' ' -f 2)
.DEFAULT_GOAL := all

.PHONY: all
all:
	make echo-commit_hash
	make app-build
	make envcheck-build
	make init
	make build

.PHONY: echo-commit_hash
echo-commit_hash:
	git rev-parse HEAD

.PHONY: app-build
app-build: 
	cd ../../provisioning/ansible/ && ./make_latest_files.sh

.PHONY: envcheck-build
envcheck-build:
	make -C ../../envcheck build

.PHONY: init
init:
	packer init -upgrade isucon13.pkr.hcl

.PHONY: build
build:
	packer build -var "commit_hash=$$(git rev-parse HEAD)" isucon13.pkr.hcl
