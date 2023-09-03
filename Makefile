
REPO=blacktop
NAME=lporg
CUR_VERSION=$(shell svu current)
NEXT_VERSION=$(shell svu patch)

.PHONY: test
test: ## Run all the tests
	@.hack/scripts/reset.sh
	@go run *.go load launchpad.yaml
	# gotestcover $(TEST_OPTIONS) -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=30s

test.save: ## Run all the tests
	@test -f launchpad.db || scripts/copy.sh
	@go run *.go save

test.default: ## Run the default test
	@.hack/scripts/reset.sh
	@go run *.go default

test.verbose: ## Run all the tests
	@.hack/scripts/reset.sh
	@go run *.go -V load launchpad.yaml

cover: test ## Run all the tests and opens the coverage report
	go tool cover -html=coverage.txt

build: ## Build a beta version of malice
	@echo "===> Building Binaries"
	go build

.PHONY: dry_release
dry_release: ## Run goreleaser without releasing/pushing artifacts to github
	@echo " > Creating Pre-release Build ${NEXT_VERSION}"
	@goreleaser build --skip-validate --id darwin --clean --single-target --output dist/lporg

.PHONY: snapshot
snapshot: ## Run goreleaser snapshot
	@echo " > Creating Snapshot ${NEXT_VERSION}"
	@goreleaser --clean --snapshot

.PHONY: release
release: ## Create a new release from the NEXT_VERSION
	@echo " > Creating Release ${NEXT_VERSION}"
	@.hack/make/release ${NEXT_VERSION}
	@goreleaser --clean

.PHONY: destroy
destroy: ## Remove release for the CUR_VERSION
	@echo " > Deleting Release"
	git tag -d ${CUR_VERSION}
	git push origin :refs/tags/${CUR_VERSION}

ci: lint test ## Run all the tests and code checks

clean: ## Clean up artifacts
	@.hack/scripts/reset.sh
	@rm -rf dist/ || true
	@rm launchpad.db || true
	@rm lporg || true

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help