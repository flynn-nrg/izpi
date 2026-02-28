#!/bin/bash
# Script to build and test BVH4 with Go 1.26 SIMD intrinsics

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "==> Testing BVH4 SIMD implementation"
echo ""
echo "Go version: $(go version)"
echo "Architecture: $(go env GOARCH)"
echo ""

cd "$PROJECT_ROOT"

# Export GOEXPERIMENT for all subsequent go commands
export GOEXPERIMENT=simd

# Run correctness tests
echo "==> Running SIMD correctness tests..."
go test ./internal/hitable -run "TestRayAABB4_SIMD" -v

# Run BVH4 tests
echo ""
echo "==> Running BVH4 integration tests..."
go test ./internal/hitable -run "TestBVH4" -v

# Benchmark comparison
echo ""
echo "==> Benchmarking SIMD vs Reference..."
go test ./internal/hitable -bench "BenchmarkRayAABB4_(SIMD|Reference)" -benchmem -benchtime=3s

# Full BVH comparison
echo ""
echo "==> Benchmarking BVH2 vs BVH4..."
go test ./internal/hitable -bench "BenchmarkBVHComparison" -benchtime=5s

echo ""
echo "==> All tests completed successfully!"
echo ""
echo "Note: On AMD64, the SIMD intrinsics version is used."
echo "      On ARM64, the pure Go loop version is used (no intrinsics yet)."
echo "      NEON intrinsics for ARM64 are expected in Go 1.27."

