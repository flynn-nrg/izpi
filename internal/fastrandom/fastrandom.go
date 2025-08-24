package fastrandom

import (
	"math/rand"
	"time"
)

type XorShift struct {
	s [2]uint64
}

func New(seed uint32) *XorShift {
	xs := &XorShift{}
	if seed == 0 {
		// Use time to get more unique initial seeds for the two uint64 states.
		t := uint64(time.Now().UnixNano())
		xs.s[0] = t
		// Combine with a shift and XOR for better differentiation of the second state.
		xs.s[1] = t ^ (t >> 32) ^ 0xbeefdead // XOR with a constant for additional entropy
	} else {
		// Use the provided initialSeed for the first state.
		xs.s[0] = uint64(seed)
		// Derive the second state from the initial seed using a constant XOR.
		xs.s[1] = uint64(seed) ^ 0xdeadbeef // A different constant XOR
	}

	// XorShift generators must not have all internal states as zero.
	// This ensures proper functioning even if the seeding somehow resulted in all zeros.
	if xs.s[0] == 0 && xs.s[1] == 0 {
		xs.s[0] = 1 // Default to 1 if both happen to be zero (highly unlikely with UnixNano)
	}

	return xs
}

func NewWithDefaults() *XorShift {
	return New(rand.Uint32())
}

// nextUint64 generates the next pseudo-random 64-bit unsigned integer
// in the sequence using the XorShift128+ algorithm.
// This is an internal helper method.
func (xs *XorShift) nextUint64() uint64 {
	s1 := xs.s[0]
	s0 := xs.s[1]

	// The XorShift128+ algorithm
	result := s0 + s1 // The sum provides better quality

	xs.s[0] = s0
	s1 ^= s1 << 23 // a
	s1 ^= s1 >> 17 // b
	s1 ^= s0       // c
	s1 ^= s0 >> 26 // d
	xs.s[1] = s1

	return result
}

// Float32 generates the next pseudo-random number in the sequence
// and returns it as a float32 within the range [0, 1).
// It does this by taking the most significant 24 bits of the generated
// uint64 and dividing by 2^24 to map it to the desired range.
func (xs *XorShift) Float32() float32 {
	// Generate a raw 64-bit random number.
	val := xs.nextUint64()

	// Extract the most significant 24 bits and divide by 2^24 (16777216.0)
	// to get a float32 in the range [0, 1).
	// Using 24 bits ensures good precision for a float32 mantissa.
	return float32(val>>40) / 16777216.0 // 16777216.0 is 2^24 as a float literal
}
