#!/usr/bin/env bash
set -e

echo "========================================="
echo " Building Multi-Target Releases..."
echo "========================================="

go run ./scripts/build.go
