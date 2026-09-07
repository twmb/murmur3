package murmur3

import (
	"hash"
	"math/bits"
)

// Make sure interfaces are correctly implemented.
var (
	_ hash.Hash   = new(digest32)
	_ hash.Hash32 = new(digest32)
	_ hash.Cloner = new(digest32)
)

const (
	c1_32 uint32 = 0xcc9e2d51
	c2_32 uint32 = 0x1b873593
)

// digest32 represents a partial evaluation of a 32 bites hash.
type digest32 struct {
	digest
	seed uint32
	h1   uint32 // Unfinalized running hash.
}

// SeedNew32 returns a hash.Hash32 for streaming 32 bit sums with its internal
// digest initialized to seed.
//
// This reads and processes the data in chunks of little endian uint32s;
// thus, the returned hash is portable across architectures.
func SeedNew32(seed uint32) hash.Hash32 {
	d := &digest32{seed: seed}
	d.bmixer = d
	d.Reset()
	return d
}

// New32 returns a hash.Hash32 for streaming 32 bit sums.
func New32() hash.Hash32 {
	return SeedNew32(0)
}

func (d *digest32) Size() int { return 4 }

func (d *digest32) reset() { d.h1 = d.seed }

func (d *digest32) Sum(b []byte) []byte {
	h := d.Sum32()
	return append(b, byte(h>>24), byte(h>>16), byte(h>>8), byte(h))
}

// Digest as many blocks as possible.
func (d *digest32) bmix(p []byte) (tail []byte) {
	h1 := d.h1
	for len(p) >= 4 {
		h1 = mix32(h1, load32(p))
		p = p[4:]
	}
	d.h1 = h1
	return p
}

func (d *digest32) Sum32() (h1 uint32) {
	return sum32(d.h1, d.tail, d.clen)
}

// Clone returns a copy of the hash. Writes to either copy do not affect the
// other. The returned hash is a hash.Hash32.
func (d *digest32) Clone() (hash.Cloner, error) {
	c := *d
	c.reseat(&c)
	return &c, nil
}

// mix32 folds one 4 byte block into the running hash.
func mix32(h1, k1 uint32) uint32 {
	k1 *= c1_32
	k1 = bits.RotateLeft32(k1, 15)
	k1 *= c2_32

	h1 ^= k1
	h1 = bits.RotateLeft32(h1, 13)
	return h1*5 + 0xe6546b64
}
