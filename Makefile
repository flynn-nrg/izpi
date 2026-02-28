.PHONY: build test bench test-simd bench-simd clean

# Default Go flags
GOFLAGS ?= 
# Enable SIMD intrinsics for optimal BVH4 performance
SIMD_FLAGS = GOEXPERIMENT=simd

# Build the izpi command
build:
	cd cmd/izpi && $(SIMD_FLAGS) go build $(GOFLAGS)

# Build without SIMD (for comparison)
build-no-simd:
	cd cmd/izpi && go build $(GOFLAGS)

# Run all tests
test:
	$(SIMD_FLAGS) go test ./...

# Run all tests without SIMD
test-no-simd:
	go test ./...

# Run SIMD-specific tests
test-simd:
	$(SIMD_FLAGS) go test ./internal/hitable -run "TestRayAABB4_SIMD" -v

# Run BVH benchmarks with SIMD enabled
bench-simd:
	$(SIMD_FLAGS) go test ./internal/hitable -bench "BenchmarkRayAABB4|BenchmarkBVHComparison" -benchmem -benchtime=3s

# Run BVH benchmarks without SIMD (for comparison)
bench-no-simd:
	go test ./internal/hitable -bench "BenchmarkRayAABB4|BenchmarkBVHComparison" -benchmem -benchtime=3s

# Compare SIMD vs Reference implementation
bench-compare:
	@echo "==> Benchmarking with SIMD intrinsics..."
	@$(SIMD_FLAGS) go test ./internal/hitable -bench "BenchmarkRayAABB4_(SIMD|Reference)" -benchmem -benchtime=3s
	@echo ""
	@echo "==> BVH2 vs BVH4 comparison..."
	@$(SIMD_FLAGS) go test ./internal/hitable -bench "BenchmarkBVHComparison" -benchtime=5s

# Full benchmark suite
bench:
	$(SIMD_FLAGS) go test ./... -bench . -benchmem

# Run benchmarks and save results for comparison
bench-save:
	@mkdir -p benchmarks
	@echo "Running benchmarks with SIMD intrinsics..."
	@$(SIMD_FLAGS) go test ./internal/hitable -bench . -benchmem -benchtime=3s | tee benchmarks/simd_$(shell date +%Y%m%d_%H%M%S).txt

# Clean build artifacts
clean:
	go clean -cache
	rm -f cmd/izpi/izpi
	rm -rf benchmarks

# Show SIMD status
info:
	@echo "Go version: $$(go version)"
	@echo "Architecture: $$(go env GOARCH)"
	@echo "SIMD Experiment: $$(GOEXPERIMENT=simd go env GOEXPERIMENT)"
	@echo ""
	@echo "BVH4 SIMD implementation:"
	@echo "  - AMD64: Uses Go 1.26 simd/archsimd (AVX2 intrinsics)"
	@echo "  - ARM64: Uses pure Go (NEON intrinsics coming in Go 1.27)"
	@echo "  - Other: Uses pure Go fallback"
	@echo ""
	@echo "To build with SIMD: make build"
	@echo "To test SIMD: make test-simd"
	@echo "To benchmark: make bench-compare"

# Help target
help:
	@echo "Available targets:"
	@echo "  build         - Build izpi with SIMD intrinsics enabled"
	@echo "  build-no-simd - Build izpi without SIMD (for comparison)"
	@echo "  test          - Run all tests with SIMD"
	@echo "  test-simd     - Run SIMD-specific tests"
	@echo "  bench-simd    - Run BVH benchmarks with SIMD"
	@echo "  bench-compare - Compare SIMD vs reference implementations"
	@echo "  bench-save    - Run benchmarks and save results"
	@echo "  info          - Show SIMD status and build info"
	@echo "  clean         - Clean build artifacts"
	@echo ""
	@echo "SIMD is enabled by default using GOEXPERIMENT=simd"

