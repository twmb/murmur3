package murmur3

import (
	"hash"
	"math/bits"
)

const (
	c1_128 = 0x87c37b91114253d5
	c2_128 = 0x4cf5ad432745937f
)

// Make sure interfaces are correctly implemented.
var (
	_ hash.Hash   = new(digest128)
	_ hash.Cloner = new(digest128)
	_ Hash128     = new(digest128)
	_ bmixer      = new(digest128)
)

// Hash128 provides an interface for a streaming 128 bit hash.
type Hash128 interface {
	hash.Hash
	Sum128() (uint64, uint64)
}

// digest128 represents a partial evaluation of a 128 bites hash.
type digest128 struct {
	digest
	seed1 uint64
	seed2 uint64
	h1    uint64 // Unfinalized running hash part 1.
	h2    uint64 // Unfinalized running hash part 2.
}

// SeedNew128 returns a Hash128 for streaming 128 bit sums with its internal
// digests initialized to seed1 and seed2.
//
// The canonical implementation allows one only uint32 seed; to imitate that
// behavior, use the same, uint32-max seed for seed1 and seed2.
func SeedNew128(seed1, seed2 uint64) Hash128 {
	d := &digest128{seed1: seed1, seed2: seed2}
	d.bmixer = d
	d.Reset()
	return d
}

// New128 returns a Hash128 for streaming 128 bit sums.
func New128() Hash128 {
	return SeedNew128(0, 0)
}

func (d *digest128) Size() int { return 16 }

func (d *digest128) reset() { d.h1, d.h2 = d.seed1, d.seed2 }

func (d *digest128) Sum(b []byte) []byte {
	h1, h2 := d.Sum128()
	return append(b,
		byte(h1>>56), byte(h1>>48), byte(h1>>40), byte(h1>>32),
		byte(h1>>24), byte(h1>>16), byte(h1>>8), byte(h1),

		byte(h2>>56), byte(h2>>48), byte(h2>>40), byte(h2>>32),
		byte(h2>>24), byte(h2>>16), byte(h2>>8), byte(h2),
	)
}

func (d *digest128) bmix(p []byte) (tail []byte) {
	h1, h2 := d.h1, d.h2
	for len(p) >= 16 {
		h1, h2 = mix128(h1, h2, load64(p), load64(p[8:]))
		p = p[16:]
	}
	d.h1, d.h2 = h1, h2
	return p
}

func (d *digest128) Sum128() (h1, h2 uint64) {
	return sum128(d.h1, d.h2, d.tail, d.clen)
}

// Clone returns a copy of the hash. Writes to either copy do not affect the
// other. The returned hash is a Hash128.
func (d *digest128) Clone() (hash.Cloner, error) {
	c := *d
	c.reseat(&c)
	return &c, nil
}

// mix128 folds one 16 byte block, already split into k1 and k2, into the
// running hash.
func mix128(h1, h2, k1, k2 uint64) (uint64, uint64) {
	k1 *= c1_128
	k1 = bits.RotateLeft64(k1, 31)
	k1 *= c2_128
	h1 ^= k1

	h1 = bits.RotateLeft64(h1, 27)
	h1 += h2
	h1 = h1*5 + 0x52dce729

	k2 *= c2_128
	k2 = bits.RotateLeft64(k2, 33)
	k2 *= c1_128
	h2 ^= k2

	h2 = bits.RotateLeft64(h2, 31)
	h2 += h1
	h2 = h2*5 + 0x38495ab5

	return h1, h2
}

func fmix64(k uint64) uint64 {
	k ^= k >> 33
	k *= 0xff51afd7ed558ccd
	k ^= k >> 33
	k *= 0xc4ceb9fe1a85ec53
	k ^= k >> 33
	return k
}
