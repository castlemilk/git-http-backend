# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o git-http-backend

# Final stage
FROM alpine:latest

# Install git and its dependencies
RUN apk add --no-cache git git-daemon

# Create directory for git repositories and set permissions
RUN mkdir -p /tmp/git && \
    chmod 777 /tmp/git

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/git-http-backend .

# Add initialization script
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Expose the default port
EXPOSE 3000

# Set environment variables
ENV GIT_PROJECT_ROOT=/tmp/git \
    GIT_HTTP_EXPORT_ALL=true \
    GIT_HTTP_MAX_REQUEST_BUFFER=1000M \
    GIT_HTTP_BACKEND_ENABLE_RECEIVE_PACK=true \
    GIT_HTTP_BACKEND_ENABLE_UPLOAD_PACK=true \
    GIT_REPO_NAME=test-repo

# Run the initialization script
ENTRYPOINT ["docker-entrypoint.sh"]