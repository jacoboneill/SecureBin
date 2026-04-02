.PHONY: clean build build/docker build/templ build/sqlc build/go test test/go test/js dev dev/templ dev/server

IN_DIR = ./cmd/server/.
OUT_DIR ?= dist
OUT_BIN = $(OUT_DIR)/securebin

# Clean
clean:
	rm -rf $(OUT_DIR) tmp static/js/node_modules

# Install deps
static/js/node_modules: static/js/package.json
	cd static/js && npm install

# Build

build: test build/templ build/sqlc build/go

build/docker:
	docker build -t securebin .

build/templ:
	go run github.com/a-h/templ/cmd/templ@latest generate

build/sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

build/go: build/sqlc build/templ
	mkdir -p $(OUT_DIR)
	go mod tidy
	CGO_ENABLED=0 go build -o $(OUT_BIN) $(IN_DIR)

# Test
test:
	@make -j2 test/go test/js

test/go: build/templ build/sqlc
	go test ./internal/handler/...

test/js: static/js/node_modules
	cd static/js && npm run test

# Dev (Server runs on localhost:7331)
dev:
	@make -j2 dev/templ dev/server

dev/templ:
	go run github.com/a-h/templ/cmd/templ@latest \
		generate --watch \
		--proxy="http://localhost:8080" \
		--open-browser=false

dev/server:
	go run github.com/air-verse/air@latest \
		--build.cmd "make build/go OUT_DIR=tmp" \
		--build.bin "tmp/securebin" \
		--build.include_ext "go,js,sql" \
		--build.exclude_regex "_templ\\.go" \
		--build.exclude_dir "tmp,internal/db" \
		--misc.clean_on_exit true
