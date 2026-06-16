# syntax=docker/dockerfile:1

# --- Stage 1: build the Svelte frontend ---------------------------------------
# Vite writes its output into ../internal/web/dist (see frontend/vite.config.ts)
# so the Go build in the next stage can pick it up via go:embed.
FROM node:22-alpine AS frontend
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN cd frontend && npm ci
COPY frontend/ ./frontend/
# internal/web must exist as the embed target; vite empties dist/ on build.
COPY internal/web/ ./internal/web/
RUN cd frontend && npm run build

# --- Stage 2: build the Go binary ---------------------------------------------
FROM golang:1.26-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
# Overlay the freshly built frontend bundle so go:embed bakes it in.
COPY --from=frontend /src/internal/web/dist/ ./internal/web/dist/
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /aivaldebatten ./cmd/server

# --- Stage 3: minimal runtime image -------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=backend /aivaldebatten /app/aivaldebatten
# config.json is read from disk at runtime (default -config config/config.json).
COPY config/ /app/config/
EXPOSE 8080
ENTRYPOINT ["/app/aivaldebatten"]
