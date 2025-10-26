#!/bin/bash
echo "Building wtop for multiple platforms..."

# Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o wtop-windows.exe wtop.go

# Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o wtop-linux wtop.go

# macOS
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o wtop-macos wtop.go

echo "Build complete!"
ls -la wtop-*
