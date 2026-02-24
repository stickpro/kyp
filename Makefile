.PHONY:
.SILENT:
.DEFAULT_GOAL := run

MIGRATIONS_DIR = ./sql/sqlite/migrations/

VERSION ?= $(strip $(shell ./scripts/version.sh))
VERSION_NUMBER := $(strip $(shell ./scripts/version.sh number))
COMMIT_HASH := $(shell git rev-parse --short HEAD)

OUT_BIN ?= ./.bin/kyp
OUT_BIN_SERVER ?= ./.bin/kypd
GO_LDFLAGS ?=
GO_OPT_BASE := -ldflags "-X main.version=$(VERSION) $(GO_LDFLAGS) -X main.commitHash=$(COMMIT_HASH)"

BUILD_ENV := CGO_ENABLED=0
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S), Linux)
	BUILD_ENV += GOOS=linux
endif
ifeq ($(UNAME_S), Darwin)
	BUILD_ENV += GOOS=darwin
endif

UNAME_P := $(shell uname -p)
ifeq ($(UNAME_P),x86_64)
	BUILD_ENV += GOARCH=amd64
endif
ifneq ($(filter arm%,$(UNAME_P)),)
	BUILD_ENV += GOARCH=arm64
endif

#build
build:
	go mod download && $(BUILD_ENV) go build $(GO_OPT_BASE) -o $(OUT_BIN) ./cmd/kyp

build-server:
	go mod download && $(BUILD_ENV) go build $(GO_OPT_BASE) -o $(OUT_BIN_SERVER) ./cmd/kypd

build-all: build build-server

run: build
	$(OUT_BIN) $(filter-out $@,$(MAKECMDGOALS))

run-server: build-server
	$(OUT_BIN_SERVER) $(filter-out $@,$(MAKECMDGOALS))
#liner
lint:
	golangci-lint run --show-stats

fmt:
	gofumpt -l -w .

# generator
gen-sql:
	cd sql && pgxgen -pgxgen-config=pgxgen.yaml -sqlc-config=sqlc.yaml crud
	cd sql && pgxgen -pgxgen-config=pgxgen.yaml -sqlc-config=sqlc.yaml sqlc generate

# Empty goals trap
%:
	@: