# Build stage — CGO_ENABLED=0
FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s \
      -X 'github.com/chickenzord/linkding-cli/internal/version.Version=${VERSION}' \
      -X 'github.com/chickenzord/linkding-cli/internal/version.GitCommit=${GIT_COMMIT}' \
      -X 'github.com/chickenzord/linkding-cli/internal/version.BuildDate=${BUILD_DATE}'" \
    -trimpath -a \
    -o linkding ./cmd/linkding

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -g 1000 -S appgroup && adduser -u 1000 -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /build/linkding .
RUN chown -R appuser:appgroup /app
USER appuser
ENTRYPOINT ["/app/linkding"]
