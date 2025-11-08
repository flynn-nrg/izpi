//go:build arm64

#ifndef BVH4_SIMD_ARM64_H
#define BVH4_SIMD_ARM64_H

#include <stdint.h>

uint8_t rayAABB4_SIMD_impl(
    const float* rayOrgX,
    const float* rayOrgY,
    const float* rayOrgZ,
    const float* rayInvDirX,
    const float* rayInvDirY,
    const float* rayInvDirZ,
    const float* minX,
    const float* minY,
    const float* minZ,
    const float* maxX,
    const float* maxY,
    const float* maxZ,
    float tMax
);

#endif // BVH4_SIMD_ARM64_H

