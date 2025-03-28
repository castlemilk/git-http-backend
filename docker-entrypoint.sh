#!/bin/sh
set -e

# Create a temporary directory for initializing the repository
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR

# Initialize git repository
git init

# Configure git user
git config --global user.email "testu@example.com"
git config --global user.name "Test User"

# Create a README.md file
echo "# ${GIT_REPO_NAME}" > README.md
echo "This is an automatically initialized repository." >> README.md

# Make initial commit
git add README.md
git commit -m "Initial commit"

# Create the bare repository
rm -rf /tmp/git/${GIT_REPO_NAME}.git
mkdir -p /tmp/git/${GIT_REPO_NAME}.git
git clone --bare $TEMP_DIR /tmp/git/${GIT_REPO_NAME}.git

# Clean up temporary directory
rm -rf $TEMP_DIR

# Set proper permissions
chmod -R 777 /tmp/git/${GIT_REPO_NAME}.git

# Start the git-http-backend server
exec /app/git-http-backend 