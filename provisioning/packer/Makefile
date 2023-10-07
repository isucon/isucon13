export BRANCH := $(shell git branch --contains | fgrep '*' | cut -d ' ' -f 2)

.PHONY: build app-build init
build: app-build init
	packer build isucon13.pkr.hcl

app-build: 
	make -C ../../webapp/go build

init:
	packer init -upgrade isucon13.pkr.hcl