//go:build amd64

#include "textflag.h"

// AVX2 implementation of rayAABB4_SIMD_impl
// Tests a ray against 4 AABBs using 256-bit YMM registers (AVX2)
//
// AVX2 provides:
// - YMM registers (256-bit = 8x float32, perfect for 4 AABBs)
// - Efficient broadcast, load, arithmetic operations
// - VBROADCASTSS for scalar replication
// - VMOVMSKPS for extracting comparison results to integer
//
// Register usage:
// - YMM0-YMM2: Ray origin (X, Y, Z) broadcasted
// - YMM3-YMM5: Ray inverse direction (X, Y, Z) broadcasted  
// - YMM6-YMM11: AABB bounds (minX, minY, minZ, maxX, maxY, maxZ)
// - YMM12-YMM15: Temporary computation registers

// func RayAABB4_SIMD(
//     rayOrgX, rayOrgY, rayOrgZ *float32,
//     rayInvDirX, rayInvDirY, rayInvDirZ *float32,
//     minX, minY, minZ *[4]float32,
//     maxX, maxY, maxZ *[4]float32,
//     tMax float32,
// ) uint8
TEXT ·RayAABB4_SIMD(SB), NOSPLIT, $0-105
    // Load argument pointers
    MOVQ    rayOrgX+0(FP), AX
    MOVQ    rayOrgY+8(FP), BX
    MOVQ    rayOrgZ+16(FP), CX
    MOVQ    rayInvDirX+24(FP), DX
    MOVQ    rayInvDirY+32(FP), SI
    MOVQ    rayInvDirZ+40(FP), DI
    MOVQ    minX+48(FP), R8
    MOVQ    minY+56(FP), R9
    MOVQ    minZ+64(FP), R10
    MOVQ    maxX+72(FP), R11
    MOVQ    maxY+80(FP), R12
    MOVQ    maxZ+88(FP), R13
    
    // Broadcast ray origin to YMM registers (lower 128 bits will have 4 copies)
    VBROADCASTSS (AX), Y0   // Y0 = {orgX, orgX, orgX, orgX, ...}
    VBROADCASTSS (BX), Y1   // Y1 = {orgY, orgY, orgY, orgY, ...}
    VBROADCASTSS (CX), Y2   // Y2 = {orgZ, orgZ, orgZ, orgZ, ...}
    
    // Broadcast ray inverse direction
    VBROADCASTSS (DX), Y3   // Y3 = {invDirX, invDirX, ...}
    VBROADCASTSS (SI), Y4   // Y4 = {invDirY, invDirY, ...}
    VBROADCASTSS (DI), Y5   // Y5 = {invDirZ, invDirZ, ...}
    
    // Load AABB bounds (4 float32 values each, 16 bytes total)
    // VMOVUPS loads 128 bits into lower half of YMM, upper half zeroed
    VMOVUPS (R8), X6        // X6 = minX[0..3] (lower 128 bits of Y6)
    VMOVUPS (R9), X7        // X7 = minY[0..3]
    VMOVUPS (R10), X8       // X8 = minZ[0..3]
    VMOVUPS (R11), X9       // X9 = maxX[0..3]
    VMOVUPS (R12), X10      // X10 = maxY[0..3]
    VMOVUPS (R13), X11      // X11 = maxZ[0..3]
    
    // ====== X AXIS ======
    // Compute t0x = (minX - orgX) * invDirX
    VSUBPS  Y0, Y6, Y12     // Y12 = minX - orgX (only lower 4 valid)
    VMULPS  Y3, Y12, Y12    // Y12 = t0x
    
    // Compute t1x = (maxX - orgX) * invDirX
    VSUBPS  Y0, Y9, Y13     // Y13 = maxX - orgX
    VMULPS  Y3, Y13, Y13    // Y13 = t1x
    
    // Get tNearX = min(t0x, t1x) and tFarX = max(t0x, t1x)
    VMINPS  Y12, Y13, Y14   // Y14 = tNearX
    VMAXPS  Y12, Y13, Y15   // Y15 = tFarX
    
    // Initialize t_min = tNearX, t_max = tFarX
    VMOVAPS Y14, Y0         // Y0 = t_min
    VMOVAPS Y15, Y1         // Y1 = t_max
    
    // ====== Y AXIS ======
    VSUBPS  Y2, Y7, Y12     // Y12 = minY - orgY (using Y2 which was orgZ, need to reload)
    
    // Reload orgY to Y2 temporarily
    VBROADCASTSS (BX), Y2
    VSUBPS  Y2, Y7, Y12     // Y12 = minY - orgY
    VMULPS  Y4, Y12, Y12    // Y12 = t0y
    
    VSUBPS  Y2, Y10, Y13    // Y13 = maxY - orgY
    VMULPS  Y4, Y13, Y13    // Y13 = t1y
    
    VMINPS  Y12, Y13, Y14   // Y14 = tNearY
    VMAXPS  Y12, Y13, Y15   // Y15 = tFarY
    
    // Update t_min = max(t_min, tNearY), t_max = min(t_max, tFarY)
    VMAXPS  Y0, Y14, Y0     // Y0 = t_min = max(t_min, tNearY)
    VMINPS  Y1, Y15, Y1     // Y1 = t_max = min(t_max, tFarY)
    
    // ====== Z AXIS ======
    VBROADCASTSS (CX), Y2   // Reload orgZ
    VSUBPS  Y2, Y8, Y12     // Y12 = minZ - orgZ
    VMULPS  Y5, Y12, Y12    // Y12 = t0z
    
    VSUBPS  Y2, Y11, Y13    // Y13 = maxZ - orgZ
    VMULPS  Y5, Y13, Y13    // Y13 = t1z
    
    VMINPS  Y12, Y13, Y14   // Y14 = tNearZ
    VMAXPS  Y12, Y13, Y15   // Y15 = tFarZ
    
    // Final t_min and t_max
    VMAXPS  Y0, Y14, Y0     // Y0 = final t_min
    VMINPS  Y1, Y15, Y1     // Y1 = final t_max
    
    // ====== CULLING CHECKS ======
    // Check 1: t_min <= t_max (equivalent to t_max >= t_min)
    VCMPPS  $0x0D, Y0, Y1, Y2   // Y2 = (t_max >= t_min) ? 0xFFFFFFFF : 0
    // VCMPPS comparison codes: 0x0D = GE (greater or equal), unordered
    
    // Check 2: t_max >= 0.0
    VXORPS  Y3, Y3, Y3          // Y3 = 0.0
    VCMPPS  $0x0D, Y3, Y1, Y3   // Y3 = (t_max >= 0) ? 0xFFFFFFFF : 0
    
    // Check 3: t_min <= tMax
    MOVSS   tMax+96(FP), X4
    VBROADCASTSS X4, Y4         // Y4 = tMax
    VCMPPS  $0x0D, Y0, Y4, Y4   // Y4 = (tMax >= t_min) ? 0xFFFFFFFF : 0
    
    // Combine all conditions with AND
    VANDPS  Y2, Y3, Y2          // Y2 = cond1 && cond2
    VANDPS  Y2, Y4, Y2          // Y2 = all conditions
    
    // Extract mask: VMOVMSKPS extracts sign bit of each float to integer
    VMOVMSKPS Y2, AX            // AX gets 8 bits (we only use lower 4)
    
    // Mask to 4 bits (we only care about lower 4 float32 values)
    ANDL    $0x0F, AX
    
    // Return as uint8
    MOVB    AX, ret+104(FP)
    
    // Clean up AVX state (required after using YMM registers)
    VZEROUPPER
    RET

