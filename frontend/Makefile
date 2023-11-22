.PHONY: all deps js_build hash build

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

all: deps build

deps:
	corepack enable
	yarn

js_build:
	yarn build

hash:
	cd $(ROOT_DIR)/dist && sha256sum `find . -type f ` > $(ROOT_DIR)/../bench/assets/data/hash.txt

build: js_build hash
