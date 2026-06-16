# aivaldebatten — build & dev tasks.
#
# Common flows:
#   make dev-backend     # run Go backend (port 8080)
#   make dev-frontend    # run Vite dev server (port 5173, proxies /ws -> 8080)
#   make build           # build frontend + single embedded binary -> bin/
#   make run             # build frontend then run the server
#   make test            # run Go tests

BINARY := bin/aivaldebatten
PKG := ./cmd/server

.PHONY: dev-backend dev-frontend build build-frontend build-backend run test fmt clean

## Development -------------------------------------------------------------

# Run the Go backend with live config. Pair with `make dev-frontend`.
dev-backend:
	go run $(PKG)

# Run the Vite dev server with hot reload; it proxies /ws to the backend.
dev-frontend:
	cd frontend && npm run dev

## Production build --------------------------------------------------------

build-frontend:
	cd frontend && npm install && npm run build

build-backend:
	mkdir -p bin
	go build -o $(BINARY) $(PKG)

# Full build: frontend first (so go:embed picks up the bundle), then binary.
build: build-frontend build-backend
	@echo "built $(BINARY)"

# Build the frontend then run the embedded single-binary server.
run: build-frontend
	go run $(PKG)

## Quality -----------------------------------------------------------------

test:
	go test ./... -race

fmt:
	gofmt -w cmd internal

clean:
	rm -rf bin
	rm -rf internal/web/dist/assets internal/web/dist/index.html
