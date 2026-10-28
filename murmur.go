// Copyright 2013, Sébastien Paolacci. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package murmur3 implements Austin Appleby's MurmurHash3 for strings and
// byte slices.
//
// Input is always read as little endian, so the sums are portable across
// architectures and match the canonical implementation on little endian
// machines.
package murmur3

// NOTE: sum32 and sum128 take a string rather than being generic. The
// compiler inlines an exported sum into a caller in another package, and the
// caller then calls sum32 or sum128 directly. When sum32 and sum128 were
// generic, the compiler moved a caller's input to the heap (issue #15). The
// []byte sums call unsafe.String themselves rather than through a helper,
// because a helper call puts Sum128 and SeedSum128 over the compiler's inline
// budget, and the compiler then emits a call for each Sum128 and SeedSum128.
// The sums only read the string from unsafe.String, and the string does not
// outlive the sum.

// bytestring lets one set of loads read both the streaming digests' []byte
// blocks and the one shot sums' string input.
type bytestring interface{ string | []byte }

// The compiler turns the shift and or chains below into a single unaligned
// load on architectures that support them, and the callers arrange for the
// bounds to already be proven so no checks remain in the loops.

func load64[T bytestring](b T) uint64 {
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

func load32[T bytestring](b T) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func load16[T bytestring](b T) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8
}

type bmixer interface {
	bmix(p []byte) (tail []byte)
	Size() (n int)
	reset()
}

type digest struct {
	clen int      // Digested input cumulative length.
	tail []byte   // 0 to Size()-1 bytes view of `buf'.
	buf  [16]byte // Expected (but not required) to be Size() large.
	bmixer
}

func (d *digest) BlockSize() int { return 1 }

func (d *digest) Write(p []byte) (n int, err error) {
	n = len(p)
	d.clen += n

	if len(d.tail) > 0 {
		// Stick back pending bytes.
		nfree := d.Size() - len(d.tail) // nfree ∈ [1, d.Size()-1].
		if nfree < len(p) {
			// One full block can be formed.
			block := append(d.tail, p[:nfree]...)
			p = p[nfree:]
			_ = d.bmix(block) // No tail.
		} else {
			// Tail's buf is large enough to prevent reallocs.
			p = append(d.tail, p...)
		}
	}

	d.tail = d.bmix(p)

	// Keep own copy of the 0 to Size()-1 pending bytes.
	nn := copy(d.buf[:], d.tail)
	d.tail = d.buf[:nn]

	return n, nil
}

func (d *digest) Reset() {
	d.clen = 0
	d.tail = nil
	d.bmixer.reset()
}

// reseat fixes the self references after a struct copy: bmixer must point at
// the copy, and tail must alias the copy's buf rather than the original's.
func (d *digest) reseat(m bmixer) {
	d.bmixer = m
	d.tail = d.buf[:len(d.tail)]
}
