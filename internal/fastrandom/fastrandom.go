package fastrandom

import (
	"math/rand"
	"sync"
)

const (
	defaultM uint64 = 4294967296
	defaultA uint64 = 1664525
	defaultC uint64 = 1013904223
)

type LCG struct {
	state uint64
	m     uint64
	a     uint64
	c     uint64

	mu sync.Mutex
}

func New(seed, m, a, c uint64) *LCG {
	return &LCG{
		state: seed,
		m:     m,
		a:     a,
		c:     c,
	}
}

func NewWithDefaults() *LCG {
	seed := rand.Uint64()

	return &LCG{
		state: seed,
		m:     defaultM,
		a:     defaultA,
		c:     defaultC,
	}
}

// Generate a random floating point number between 0 and 1.
func (l *LCG) Float64() float64 {

	/* Update the LCG state using the formula Xn+1 = (A*Xn + C) mod M */
	l.mu.Lock()
	defer l.mu.Unlock()
	l.state = (l.a*l.state + l.c) % l.m
	/* Convert the LCG state to a floating point number between 0 and 1 */
	return float64(l.state) / float64(l.m)
}
