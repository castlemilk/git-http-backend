#!/bin/sh
set -e

# Create the repository directory
mkdir -p /tmp/git/${GIT_REPO_NAME}.git
cd /tmp/git/${GIT_REPO_NAME}.git
git init --bare

# Set proper permissions
chmod -R 777 /tmp/git/${GIT_REPO_NAME}.git

# Start the git-http-backend server
exec /app/git-http-backend 