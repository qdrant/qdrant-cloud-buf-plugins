.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9\/-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt and gci against code.
	go fmt ./...
	$(GCI) write ./ --skip-generated -s standard -s default -s 'prefix(github.com/qdrant)' -s 'prefix(github.com/qdrant/qdrant-cloud-buf-plugins/)'

.PHONY: vet
vet: ## Run go vet (static analysis tool) against code.
	go vet ./...

.PHONY: lint
lint: bootstrap ## Run project linters
	$(GOLANGCI_LINT) run

.PHONY: test
test: fmt vet test_unit ## Run all tests.

.PHONY: test_unit
test_unit: ## Run unit tests.
	echo "Running Go unit tests..."
	go test ./... -coverprofile cover.out -v

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GO ?= go
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
GCI = $(LOCALBIN)/gci

## Tool Versions
GOLANGCI_LINT_VERSION ?= v2.6.1
GCI_VERSION ?= v0.13.5

.PHONY: bootstrap
bootstrap: install/gci install/golangci-lint ## Install required dependencies to work with this project

.PHONY: golangci-lint
install/golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: install/gci
install/gci: $(GCI) ## Download gci locally if necessary.
$(GCI): $(LOCALBIN)
	$(call go-install-tool,$(GCI),github.com/daixiang0/gci,$(GCI_VERSION))

# copied from kube-builder
# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef
