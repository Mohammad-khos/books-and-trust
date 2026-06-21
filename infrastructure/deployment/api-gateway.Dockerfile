# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.25
ARG ALPINE_VERSION=3.23

# =====================
# Stage 1: deps
# =====================

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS deps

WORKDIR /app



# Copy go.mod and go.sum first to leverage Docker cache for dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download -x


FROM deps AS builder


ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY services/api-gateway/ ./services/api-gateway/
COPY shared/ ./shared/ 

WORKDIR /app/services/api-gateway
RUN --mount=type=cache,target=/root/.cache/go-build \
     --mount=type=cache,target=/go/pkg/mod \
     go build \
    -trimpath \
    -ldflags="-w -s" \
    -o /app/api-gateway ./cmd/api

# =====================
# Stage 2: Runtime Image
# =====================
FROM scratch AS runtime

COPY --from=builder /app/api-gateway /app/api-gateway

WORKDIR /app


EXPOSE 8081

# Command to run the executable
# The binary is now at /app/api-gateway
ENTRYPOINT ["/app/api-gateway"]
