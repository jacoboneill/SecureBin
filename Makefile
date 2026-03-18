.PHONY: clean install test test-go test-js build build-docker dev

REQUIRED_BINS := go npm docker sqlc templ
$(foreach bin,$(REQUIRED_BINS),\
  $(if $(shell command -v $(bin) 2>/dev/null),,\
    $(error "$(bin) is not installed. Please install it before running make")))

# Directories
DIST_DIR := dist
GO_BIN := $(DIST_DIR)/securebin

clean:
	rm -rf $(DIST_DIR)
	rm -rf static/js/node_modules

static/js/node_modules: static/js/package.json
	cd static/js && npm install

test: test-go test-js

test-go:
	go test ./...

test-js: static/js/node_modules
	cd static/js && npm run test

build: test
	mkdir -p $(DIST_DIR)
	sqlc generate
	templ generate
	CGO_ENABLED=0 go build -o $(GO_BIN) ./cmd/server/.

build-docker: test
	docker build -t securebin .

dev:
	air
