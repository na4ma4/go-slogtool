APP_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_HASH ?= $(shell git show -s --format=%h)

GO_DEBUG_ARGS   ?= -v -ldflags "-X main.version=$(GO_APP_VERSION)+debug -X main.commit=$(GIT_HASH) -X main.date=$(APP_DATE) -X main.builtBy=makefiles -X main.repo=$(GIT_SLUG)"
GO_RELEASE_ARGS ?= -v -ldflags "-X main.version=$(GO_APP_VERSION) -X main.commit=$(GIT_HASH) -X main.date=$(APP_DATE) -X main.builtBy=makefiles -X main.repo=$(GIT_SLUG) -s -w"

-include .makefiles/Makefile
-include .makefiles/ext/cwx/Makefile
-include .makefiles/pkg/go/v1/Makefile
-include .makefiles/ext/cwx/pkg/go/v1/Makefile
-include .makefiles/ext/na4ma4/lib/golangci-lint/v1/Makefile

.makefiles/ext/cwx/%: .makefiles/Makefile
	@curl -sfL https://makefiles.cwx.io/v1 | bash /dev/stdin "$@"

.makefiles/ext/na4ma4/%: .makefiles/Makefile
	@curl -sfL https://raw.githubusercontent.com/na4ma4/makefiles-ext/main/v1/install | bash /dev/stdin "$@"

.makefiles/%:
	@curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"


######################
# Linting
######################

ci:: lint
