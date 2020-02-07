package murmur3

const (
	c1 uint32 = 0xcc9e2d51
	c2 uint32 = 0x1b873593
	c3 uint32 = 0x85ebca6b
	c4 uint32 = 0xc2b2ae35
	c5 uint32 = 0xe6546b64
)

// StringSum32 is the string version of Sum32.
func StringSum32(data string) uint32 {
	return SeedSum32(0, quickslice(data))
}

// SeedStringSum32 is the string version of SeedSum32.
func SeedStringSum32(seed uint32, data string) (h1 uint32) {
	return SeedSum32(seed, quickslice(data))
}

// Sum32 returns the murmur3 sum of data. It is equivalent to the following
// sequence (without the extra burden and the extra allocation):
//     hasher := New32()
//     hasher.Write(data)
//     return hasher.Sum32()
func Sum32(data []byte) uint32 {
	return SeedSum32(0, data)
}

// SeedSum32 returns the murmur3 sum of data with the digest initialized to
// seed.
func SeedSum32(seed uint32, data []byte) (h1 uint32) {
	h1 = seed
	dataLen := len(data)

	var k uint32
	for i := dataLen >> 2; i != 0; i-- {
		k = uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
		data = data[4:]
		h1 ^= mix(k)
		h1 = (h1 << 13) | (h1 >> 19)
		h1 = h1*5 + c5
	}

	k = 0
	for i := dataLen & 3; i != 0; i-- {
		k <<= 8
		k |= uint32(data[i-1])
	}

	h1 ^= mix(k)

	h1 ^= uint32(dataLen)
	h1 ^= h1 >> 16
	h1 *= c3
	h1 ^= h1 >> 13
	h1 *= c4
	h1 ^= h1 >> 16

	return h1
}

func mix(k uint32) uint32 {
	k *= c1
	k = (k << 15) | (k >> 17)
	k *= c2
	return k
}
