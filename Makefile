# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

# Setup name variables for the package/tool
NAME := cni-benchmarks
PKG := github.com/jessfraz/$(NAME)

# Set any default go build tags
BUILDTAGS :=

# Set the build dir, where built cross-compiled binaries will be output
BUILDDIR := ${PREFIX}/cross
BINDIR := ${PREFIX}/bin

# Populate version variables
# Add to compile time flags
VERSION := $(shell cat VERSION.txt)
GITCOMMIT := $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
	GITCOMMIT := $(GITCOMMIT)-dirty
endif
CTIMEVAR=-X $(PKG)/version.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/version.VERSION=$(VERSION)
GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

# List the GOOS and GOARCH to build
GOOSARCHES = linux/arm linux/arm64 linux/amd64 linux/386

.PHONY: build
build: $(NAME) ## Builds a dynamic executable or package

$(NAME): *.go VERSION.txt
	@echo "+ $@"
	go build -tags "$(BUILDTAGS)" ${GO_LDFLAGS} -o $(NAME) .

.PHONY: static
static: ## Builds a static executable
	@echo "+ $@"
	CGO_ENABLED=0 go build \
				-tags "$(BUILDTAGS) static_build" \
				${GO_LDFLAGS_STATIC} -o $(NAME) .

all: clean build fmt lint test staticcheck vet install ## Runs a clean, build, fmt, lint, test, staticcheck, vet and install

.PHONY: fmt
fmt: ## Verifies all files have men `gofmt`ed
	@echo "+ $@"
	@gofmt -s -l . | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: lint
lint: ## Verifies `golint` passes
	@echo "+ $@"
	@golint ./... | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: test
test: ## Runs the go tests
	@echo "+ $@"
	@go test -v -tags "$(BUILDTAGS) cgo" $(shell go list ./... | grep -v vendor)

.PHONY: vet
vet: ## Verifies `go vet` passes
	@echo "+ $@"
	@go vet $(shell go list ./... | grep -v vendor) | grep -v '.pb.go:' | tee /dev/stderr

.PHONY: staticcheck
staticcheck: ## Verifies `staticcheck` passes
	@echo "+ $@"
	@staticcheck $(shell go list ./... | grep -v vendor) | grep -v '.pb.go:' | tee /dev/stderr

.PHONY: cover
cover: ## Runs go test with coverage
	@echo "" > coverage.txt
	@for d in $(shell go list ./... | grep -v vendor); do \
		go test -race -coverprofile=profile.out -covermode=atomic "$$d"; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi; \
	done;

.PHONY: install
install: ## Installs the executable or package
	@echo "+ $@"
	go install -a -tags "$(BUILDTAGS)" ${GO_LDFLAGS} .

define buildpretty
mkdir -p $(BUILDDIR)/$(1)/$(2);
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 go build \
	 -o $(BUILDDIR)/$(1)/$(2)/$(NAME) \
	 -a -tags "$(BUILDTAGS) static_build netgo" \
	 -installsuffix netgo ${GO_LDFLAGS_STATIC} .;
md5sum $(BUILDDIR)/$(1)/$(2)/$(NAME) > $(BUILDDIR)/$(1)/$(2)/$(NAME).md5;
sha256sum $(BUILDDIR)/$(1)/$(2)/$(NAME) > $(BUILDDIR)/$(1)/$(2)/$(NAME).sha256;
endef

.PHONY: cross
cross: *.go VERSION.txt ## Builds the cross-compiled binaries, creating a clean directory structure (eg. GOOS/GOARCH/binary)
	@echo "+ $@"
	$(foreach GOOSARCH,$(GOOSARCHES), $(call buildpretty,$(subst /,,$(dir $(GOOSARCH))),$(notdir $(GOOSARCH))))

define buildrelease
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 go build \
	 -o $(BUILDDIR)/$(NAME)-$(1)-$(2) \
	 -a -tags "$(BUILDTAGS) static_build netgo" \
	 -installsuffix netgo ${GO_LDFLAGS_STATIC} .;
md5sum $(BUILDDIR)/$(NAME)-$(1)-$(2) > $(BUILDDIR)/$(NAME)-$(1)-$(2).md5;
sha256sum $(BUILDDIR)/$(NAME)-$(1)-$(2) > $(BUILDDIR)/$(NAME)-$(1)-$(2).sha256;
endef

.PHONY: release
release: *.go VERSION.txt ## Builds the cross-compiled binaries, naming them in such a way for release (eg. binary-GOOS-GOARCH)
	@echo "+ $@"
	$(foreach GOOSARCH,$(GOOSARCHES), $(call buildrelease,$(subst /,,$(dir $(GOOSARCH))),$(notdir $(GOOSARCH))))

.PHONY: bump-version
BUMP := patch
bump-version: ## Bump the version in the version file. Set BUMP to [ patch | major | minor ]
	@go get -u github.com/jessfraz/junk/sembump # update sembump tool
	$(eval NEW_VERSION = $(shell sembump --kind $(BUMP) $(VERSION)))
	@echo "Bumping VERSION.txt from $(VERSION) to $(NEW_VERSION)"
	echo $(NEW_VERSION) > VERSION.txt
	@echo "Updating links to download binaries in README.md"
	sed -i s/$(VERSION)/$(NEW_VERSION)/g README.md
	git add VERSION.txt README.md
	git commit -vsam "Bump version to $(NEW_VERSION)"
	@echo "Run make tag to create and push the tag for new version $(NEW_VERSION)"

.PHONY: tag
tag: ## Create a new git tag to prepare to build a release
	git tag -sa $(VERSION) -m "$(VERSION)"
	@echo "Run git push origin $(VERSION) to push your new tag to GitHub and trigger a travis build."

.PHONY: AUTHORS
AUTHORS:
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@echo "$(shell git log --format='\n%aN <%aE>' | LC_ALL=C.UTF-8 sort -uf)" >> $@

DOCKER_DEV_IMAGE=r.j3ss.co/cni-benchmarks-dev
.PHONY: build-dev-image
build-dev-image:
	@docker build --rm --force-rm -t $(DOCKER_DEV_IMAGE) -f Dockerfile.dev .

.PHONY: update-binaries
update-binaries: clean-binaries build-dev-image ## Run the dev dockerfile which builds all the cni binaries for testing.
	-$(shell docker run --rm --disable-content-trust=true $(DOCKER_DEV_IMAGE) bash -c 'tar -c /cni/bin' | tar -xv --strip-components=1 -C . > /dev/null)

LOCAL_IP_ENV?=$(shell ip route get 8.8.8.8 | head -1 | awk '{print $$7}')

.PHONY: run-containers
run-containers: stop-containers clean run-etcd run-calico run-cilium run-flannel run-weave ## Runs the etcd, calico, cilium, flannel, and weave containers.

.PHONY: stop-containers
stop-containers: stop-etcd stop-calico stop-cilium stop-flannel stop-weave ## Stops all the running containers.

ETCD_CONTAINER_NAME=cni-etcd
ETCD_ENDPOINTS=http://127.0.0.1:2379
.PHONY: run-etcd
run-etcd: stop-etcd ## Run etcd in a container for testing calico and cilium against.
	docker run --detach \
		--restart always \
		-p 127.0.0.1:2379:2379 \
		-v /tmp/etcd:/etcd-data \
		--name $(ETCD_CONTAINER_NAME) \
		quay.io/coreos/etcd \
			etcd \
			--data-dir=/etcd-data \
			--advertise-client-urls "http://$(LOCAL_IP_ENV):2379,$(ETCD_ENDPOINTS)" \
			--listen-client-urls "http://0.0.0.0:2379"

.PHONY: stop-etcd
stop-etcd: # Stops the etcd container.
	@-docker rm -f $(ETCD_CONTAINER_NAME)

CALICO_CONTAINER_NAME=cni-calico
.PHONY: clean run-calico
run-calico: stop-calico run-etcd ## Run calico in a container for testing calico against.
	docker run --detach \
		--restart always \
		-e "ETCD_ENDPOINTS=$(ETCD_ENDPOINTS)" \
		-e "CALICO_IP=autodetect" \
		-v /var/lib/calico:/var/lib/calico \
		-v /var/run/calico:/var/run/calico \
		-v /lib/modules:/lib/modules \
		-v /tmp/calico:/var/log/calico \
		--privileged \
		--net host \
		--name $(CALICO_CONTAINER_NAME) \
		quay.io/calico/node

.PHONY: stop-calico
stop-calico: # Stops the calico container.
	@-docker rm -f $(CALICO_CONTAINER_NAME)

CILIUM_CONTAINER_NAME=cni-cilium
.PHONY: run-cilium
run-cilium: stop-cilium run-etcd ## Run cilium in a container for testing cilium against.
	docker run --detach \
		--restart always \
		-v /sys/fs/bpf:/sys/fs/bpf \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /var/run/cilium:/var/run/cilium \
		--privileged \
		--net host \
		--name $(CILIUM_CONTAINER_NAME) \
		cilium/cilium \
		cilium-agent \
			-t=vxlan \
			--kvstore=etcd \
			--kvstore-opt=etcd.address=$(ETCD_ENDPOINTS)

.PHONY: stop-cilium
stop-cilium: # Stops the cilium container.
	@-docker rm -f $(CILIUM_CONTAINER_NAME)

FLANNEL_CONTAINER_NAME=cni-flannel
FLANNEL_VERSION=v0.10.0-amd64
.PHONY: run-flannel
run-flannel: stop-flannel run-etcd ## Run flannel in a container for testing flannel against.
	docker run --detach \
		--restart always \
		-v /var/run/flannel:/var/run/flannel \
		--privileged \
		--net host \
		--name $(FLANNEL_CONTAINER_NAME) \
		quay.io/coreos/flannel:$(FLANNEL_VERSION) \
			--ip-masq \
			--etcd-endpoints=$(ETCD_ENDPOINTS)
	-docker run --rm \
		--net=host \
		quay.io/coreos/etcd \
			etcdctl \
			set /coreos.com/network/config \
			'{ "Network": "10.6.0.0/16", "Backend": {"Type": "vxlan"}}'

.PHONY: stop-flannel
stop-flannel: # Stops the flannel container.
	@-docker rm -f $(FLANNEL_CONTAINER_NAME)
	@-sudo ip link delete flannel.1

.PHONY: run-weave
run-weave: stop-weave ## Run weave in a container for testing weave against.
	docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(shell which docker):/usr/bin/docker:ro \
		-v /tmp/weave:/weavedb \
		--privileged \
		--net host \
		weaveworks/weaveexec launch

.PHONY: stop-weave
stop-weave: # Stops the weave containers.
	@-docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(shell which docker):/usr/bin/docker:ro \
		-v /tmp/weave:/weavedb \
		--privileged \
		--net host \
		weaveworks/weaveexec stop
	@-docker rm -f weave weavedb weavevolumes-2.3.0

.PHONY: clean-binaries
clean-binaries:
	$(RM) -r $(BINDIR)

.PHONY: clean
clean: stop-containers ## Cleanup any build binaries or packages
	@echo "+ $@"
	$(RM) $(NAME)
	$(RM) -r $(BUILDDIR)
	sudo $(RM) -r /var/lib/calico /tmp/etcd /tmp/calico /tmp/weave

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
