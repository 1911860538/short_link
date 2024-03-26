GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
EXCLUDED_TESTS_DIR := $(shell go list ./... | grep -v '/tests')

.PHONY: all
all: git_branch go_vet go_mod go_test check_deadcode go_fmt

.PHONY: git_branch
git_branch:
	@echo "current git branch is" $(GIT_BRANCH)


.PHONY: go_vet
go_vet:
	@echo "run go vet"
	@go vet ./...

.PHONY: go_mod
go_mod:
	@echo "run go mod"
	@go mod download
	@go mod tidy

.PHONY: go_test
go_test:
	@echo "run go test"
	@go test -cover ./...


.PHONY: check_deadcode
check_deadcode:
	@echo "run check_deadcode"
	@go install golang.org/x/tools/cmd/deadcode@latest
	@deadcode $(EXCLUDED_TESTS_DIR)


.PHONY: go_fmt
go_fmt:
	@echo "run go fmt"
	@go fmt ./...
