# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
BINARY_NAME=git-http-backend

# Docker parameters
DOCKER_IMAGE=git-http-backend
DOCKER_TAG=latest
DOCKER_PORT=3000
DOCKER_VOLUME=/tmp/git-repos
REPO ?= test-repo

# Build flags
BUILD_FLAGS=-v

.PHONY: all build test clean docker-build docker-run docker-stop help

all: test build

help:
	@echo "Available targets:"
	@echo "  all        - Run tests and build binary"
	@echo "  build      - Build binary"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container (use REPO=<name> to specify repository name)"
	@echo "  docker-stop  - Stop and remove Docker container"

build:
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME)

test:
	$(GOTEST) -v ./...

clean:
	rm -f $(BINARY_NAME)
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true

docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-stop
	docker run -d \
		--name $(BINARY_NAME) \
		-p $(DOCKER_PORT):$(DOCKER_PORT) \
		-v $(DOCKER_VOLUME):/tmp/git \
		-e GIT_HTTP_BACKEND_ENABLE_RECEIVE_PACK=true \
		-e GIT_HTTP_BACKEND_ENABLE_UPLOAD_PACK=true \
		-e GIT_HTTP_EXPORT_ALL=true \
		-e GIT_REPO_NAME=$(REPO) \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop:
	docker stop $(BINARY_NAME) 2>/dev/null || true
	docker rm $(BINARY_NAME) 2>/dev/null || true
