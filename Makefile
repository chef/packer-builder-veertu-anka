LATEST-GIT-SHA := $(shell git rev-parse HEAD)
VERSION := $(shell cat VERSION)
FLAGS := -X main.commit=$(LATEST-GIT-SHA) -X main.version=$(VERSION)
BIN := packer-plugin-veertu-anka
SOURCES := $(shell find . -name '*.go')

.PHONY: test packer-test clean clean-images

test:
	go test -v builder/ankavm/*.go
	go test -v post-processor/ankaregistry/*.go

hcl2spec:
	GOOS=darwin GOBIN=$(shell pwd) go install github.com/hashicorp/packer/cmd/mapstructure-to-hcl2
	GOOS=darwin PATH="$(shell pwd):${PATH}" go generate builder/ankavm/config.go
	GOOS=darwin PATH="$(shell pwd):${PATH}" go generate post-processor/ankaregistry/post-processor.go
	GOOS=darwin PATH="$(shell pwd):${PATH}" go generate post-processor/ankastop/post-processor.go

build: $(BIN)
$(BIN): hcl2spec
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

packer-test-push: build
	PACKER_LOG=1 packer build examples/macos-mojave-registry-push.json

clean:
	rm -f $(BIN)
	rm -f mapstructure-to-hcl2

clean-images:
	anka --machine-readable list | jq -r '.body[].name' | grep anka-packer | xargs -n1 anka delete --yes

clean-clones:
	anka --machine-readable list | jq -r '.body[].name' | grep anka-packer | grep -v base | xargs -n1 anka delete --yes

wipe-anka:
	-rm -rf ~/Library/Application\ Support/Veertu
	-rm -rf ~/.anka