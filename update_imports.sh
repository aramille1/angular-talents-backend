#!/bin/bash

echo "Updating import paths in Go files..."

# Find all Go files and update imports
find . -name "*.go" | xargs sed -i '' 's/reverse-job-board/angular-talents-backend/g'

echo "All import paths updated!"
