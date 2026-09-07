package murmur3

import "math/bits"

// Sum32 returns the murmur3 sum of data. It is equivalent to the following
// sequence (without the extra burden and the extra allocation):
//
//	hasher := New32()
//	hasher.Write(data)
//	return hasher.Sum32()
func Sum32(data []byte) uint32 {
	return sum32(0, data, len(data))
}

// SeedSum32 returns the murmur3 sum of data with the digest initialized to
// seed.
//
// This reads and processes the data in chunks of little endian uint32s;
// thus, the returned hash is portable across architectures.
func SeedSum32(seed uint32, data []byte) (h1 uint32) {
	return sum32(seed, data, len(data))
}

// StringSum32 is the string version of Sum32.
func StringSum32(data string) uint32 {
	return sum32(0, data, len(data))
}

// SeedStringSum32 is the string version of SeedSum32.
func SeedStringSum32(seed uint32, data string) (h1 uint32) {
	return sum32(seed, data, len(data))
}

// sum32 mixes all of data into the running h1 and then finalizes with clen,
// the total number of bytes hashed. The one shot sums pass all of their
// input; the streaming digest passes its leftover tail.
func sum32[T bytestring](h1 uint32, data T, clen int) uint32 {
	// Strictly greater for the same reason as in sum128: it removes the
	// compiler's zero length pointer guard from the loop.
	for len(data) > 4 {
		h1 = mix32(h1, load32(data))
		data = data[4:]
	}
	if len(data) >= 4 {
		h1 = mix32(h1, load32(data))
		data = data[4:]
	}

	var k1 uint32
	switch len(data) {
	case 3:
		k1 ^= uint32(data[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(data[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(data[0])
		k1 *= c1_32
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2_32
		h1 ^= k1
	}

	h1 ^= uint32(clen)

	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return h1
}
