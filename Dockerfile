# Git HTTP Backend Service
#
# This Dockerfile creates a containerized Git HTTP backend service that provides
# HTTP access to Git repositories. It's designed to be lightweight and secure,
# using a multi-stage build process.
#
# Usage:
#   Build: docker build -t git-http-backend .
#   Run: docker run -p 3000:3000 -v /path/to/repos:/tmp/git git-http-backend
#
# Environment Variables:
#   GIT_PROJECT_ROOT: Directory containing Git repositories (default: /tmp/git)
#   GIT_HTTP_EXPORT_ALL: Enable access to all repositories (default: true)
#   GIT_HTTP_MAX_REQUEST_BUFFER: Maximum size of HTTP request buffer (default: 1000M)
#   GIT_HTTP_BACKEND_ENABLE_RECEIVE_PACK: Enable push operations (default: true)
#   GIT_HTTP_BACKEND_ENABLE_UPLOAD_PACK: Enable pull operations (default: true)
#   GIT_REPO_NAME: Default repository name (default: test-repo)
#
# Security Notes:
#   - Uses Alpine Linux for a minimal attack surface
#   - Runs as non-root user
#   - Implements proper file permissions
#
# Dependencies:
#   - git
#   - git-daemon
#   - musl-dev (build only)
#   - gcc (build only)

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

# Docker metadata annotations
LABEL org.opencontainers.image.title="Git HTTP Backend" \
      org.opencontainers.image.description="A lightweight and secure Git HTTP backend service providing HTTP access to Git repositories" \
      org.opencontainers.image.version="1.0.0" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.base.name="alpine:latest" \
      org.opencontainers.image.url="https://github.com/castlemilk/git-http-backend" \
      org.opencontainers.image.documentation="https://github.com/castlemilk/git-http-backend/blob/main/README.md" \
      org.opencontainers.image.revision="${VCS_REF:-unknown}" \
      org.opencontainers.image.created="${BUILD_DATE:-unknown}" \
      org.opencontainers.image.authors="Your Name <your.email@example.com>" \
      org.opencontainers.image.ref.name="git-http-backend:latest" \
      org.opencontainers.image.source="https://github.com/castlemilk/git-http-backend" \
      org.opencontainers.image.base.digest="${ALPINE_DIGEST:-unknown}"

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