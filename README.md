# Git HTTP Backend Server

A simple Git HTTP backend server with basic authentication support. This server allows you to host Git repositories over HTTP with password protection.

## Features

- Basic authentication support
- Git receive-pack and upload-pack support
- Docker support
- Configurable port and repository directory
- Environment variable configuration
- Configuration via environment variables, command-line flags, or config files

## Prerequisites

- Go 1.24 or later
- Git
- Docker (optional)

## Building

### Local Build

```bash
# Build the binary
make build

# Run tests
make test

# Build and test
make all
```

### Docker Build

```bash
# Build the Docker image
make docker-build

# Run the Docker container
make docker-run

# Stop the Docker container
make docker-stop
```

## Configuration

The server can be configured using multiple methods (in order of precedence):

1. Environment variables (prefixed with `GIT_`)
2. Command-line flags
3. Configuration file (config.yaml, config.json, etc.)
4. Default values

### Available Configuration Options

| Option | Environment Variable | Default Value | Description |
|--------|---------------------|---------------|-------------|
| Port | `GIT_PORT` | 3000 | Port to listen on |
| Server Temp Dir | `GIT_SERVER_TEMP_DIR` | /tmp/git | Directory to store Git repositories |
| Username | `GIT_USERNAME` | testuser | Username for basic auth |
| Password | `GIT_PASSWORD` | testpass | Password for basic auth |

### Example Usage

```bash
# Using environment variables
export GIT_PORT=8080
export GIT_USERNAME=admin
export GIT_PASSWORD=secret
./git-http-backend

# Using command-line flags
./git-http-backend --port 8080 --username admin --password secret

# Using Docker with environment variables
docker run -e GIT_PORT=8080 -e GIT_USERNAME=admin -e GIT_PASSWORD=secret git-http-backend
```

## Usage

### Running the Server

```bash
# Run locally
./git-http-backend --port 3000 --server-temp-dir /path/to/repos

# Run with Docker
make docker-run
```

### Using with Git

```bash
# Clone a repository
git clone http://localhost:3000/repo.git

# Push to a repository
git push http://localhost:3000/repo.git main

# When prompted, use the configured username and password
```

## Development

### Running Tests

```bash
make test
```

### Clean Build Artifacts

```bash
make clean
```

## License

MIT License 