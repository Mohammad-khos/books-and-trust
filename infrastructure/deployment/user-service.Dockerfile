# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.25
ARG ALPINE_VERSION=3.23

# =====================
# Stage 1: deps
# =====================

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS deps

WORKDIR /app

#set mirror
RUN go env -w GOPROXY=https://go.devneeds.ir,direct
RUN go env -w GOSUMDB="sum.golang.org https://go-sum.devneeds.ir"

# Copy go.mod and go.sum first to leverage Docker cache for dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download -x


FROM deps AS builder


ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY services/user-service/ ./services/user-service/
COPY shared/ ./shared/ 


WORKDIR /app/services/user-service
RUN --mount=type=cache,target=/root/.cache/go-build \
     --mount=type=cache,target=/go/pkg/mod \
     go build \
    -trimpath \
    -ldflags="-w -s" \
    -o /app/user-service ./cmd/api

# =====================
# Stage 2: Runtime Image
# =====================
FROM scratch AS runtime

COPY --from=builder /app/user-service /app/user-service

WORKDIR /app

# Command to run the executable
# The binary is now at /app/user-service
ENTRYPOINT ["/app/user-service"]
