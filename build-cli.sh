#!/bin/bash

# Create .binaries folder if it doesn't exist
mkdir -p .binaries

# Build the CLI
go build -o .binaries/cli ./cmd/cli