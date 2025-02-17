LATEST-GIT-SHA := $(shell git rev-parse HEAD)
VERSION := $(shell cat VERSION)
FLAGS := -X main.commit=$(LATEST-GIT-SHA) -X main.version=$(VERSION)
BIN := packer-builder-veertu-anka
SOURCES := $(shell find . -name '*.go')

.PHONY: test packer-test clean clean-images

test:
	go test -v builder/anka/*.go

build: $(BIN)
$(BIN):
	GOOS=darwin GOBIN=$(shell pwd) go install github.com/hashicorp/packer/cmd/mapstructure-to-hcl2
	GOOS=darwin PATH="$(shell pwd):${PATH}" go generate builder/anka/config.go
	GOOS=darwin go build -ldflags="$(FLAGS)" -o $(BIN)

install: $(BIN)
	mkdir -p ~/.packer.d/plugins/
	cp $(BIN) ~/.packer.d/plugins/

build-and-install: $(BIN)
	$(MAKE) clean
	$(MAKE) build
	$(MAKE) install

packer-test: install
	PACKER_LOG=1 packer build examples/create-from-installer.json

clean:
	rm -f $(BIN)

clean-images:
	anka --machine-readable list | jq -r '.body[].name' | grep anka-packer | xargs -n1 anka delete --yes

clean-clones:
	anka --machine-readable list | jq -r '.body[].name' | grep anka-packer | grep -v base | xargs -n1 anka delete --yes

wipe-anka:
	-rm -rf ~/Library/Application\ Support/Veertu
	-rm -rf ~/.anka