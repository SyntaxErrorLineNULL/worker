# This Makefile defines common tasks for a Go project, including linting the code,
# running tests, and generating mocks using mockery.

# Target: lint
# Description: Run the Go linter using golangci-lint.
# This checks the code for stylistic issues, potential bugs, and other improvements.
lint: ## Run linter
	golangci-lint run

# Target: test
# Description: Run all tests in the project.
# The -v flag makes the test output verbose, showing detailed information about each test.
test: ## Run tests
	go test -v ./...

# Target: generate
# Description: Generate code, particularly mocks using mockery.
# If mockery is not installed, this will install it first, and then run go generate.
generate: ## Generate mocks
# Check if mockery is installed. If not, install it.
ifeq (, $(shell which mockery))
	go install github.com/vektra/mockery/v2@v2.44.2
endif
	# Run go generate to trigger code generation in the project.
	go generate ./...

semver-cli:
ifeq (, $(shell which semver-cli))
	@printf "\033[36m%s\033[0m\n" "Installing semvercli..."
	go install github.com/jfwenisch/semver-cli@latest
endif

bump: semver-cli ## Bump version, generate git tag and push it to repository
	@printf "\033[36m%s\033[0m\n" "Bumping version..."
	@git fetch --all --tags
	@semver-cli tags bump -t patch -p v