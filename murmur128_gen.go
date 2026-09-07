package murmur3

import "math/bits"

// Sum128 returns the murmur3 sum of data. It is equivalent to the following
// sequence (without the extra burden and the extra allocation):
//
//	hasher := New128()
//	hasher.Write(data)
//	return hasher.Sum128()
func Sum128(data []byte) (h1 uint64, h2 uint64) {
	return sum128(0, 0, data, len(data))
}

// SeedSum128 returns the murmur3 sum of data with digests initialized to seed1
// and seed2.
//
// The canonical implementation allows only one uint32 seed; to imitate that
// behavior, use the same, uint32-max seed for seed1 and seed2.
//
// This reads and processes the data in chunks of little endian uint64s;
// thus, the returned hashes are portable across architectures.
func SeedSum128(seed1, seed2 uint64, data []byte) (h1 uint64, h2 uint64) {
	return sum128(seed1, seed2, data, len(data))
}

// StringSum128 is the string version of Sum128.
func StringSum128(data string) (h1 uint64, h2 uint64) {
	return sum128(0, 0, data, len(data))
}

// SeedStringSum128 is the string version of SeedSum128.
func SeedStringSum128(seed1, seed2 uint64, data string) (h1 uint64, h2 uint64) {
	return sum128(seed1, seed2, data, len(data))
}

// sum128 mixes all of data into the running h1 and h2 and then finalizes
// with clen, the total number of bytes hashed. The one shot sums pass all
// of their input; the streaming digest passes its leftover tail.
func sum128[T bytestring](h1, h2 uint64, data T, clen int) (uint64, uint64) {
	if len(data) >= 16 {
		// With at least one full block, the last 16 bytes of the input are
		// always in bounds, and the tail is whatever of them the block loop
		// will not consume. Read it out of them here with two overlapping
		// loads instead of branching on its length afterwards: a switch on
		// the length is a jump table, and on varying key lengths it
		// mispredicts nearly every time. Doing it before the loop also
		// leaves only two words live across the loop and hides the loads
		// behind it. k1 is the first eight tail bytes, loaded exactly when
		// n >= 8 and otherwise shifted down out of the last eight bytes;
		// k2 is whatever sits above those. A shift of 64 or more is zero
		// in Go, so k2 is zero for a tail of eight bytes or fewer, and
		// mixing zero changes nothing, so the mixes need no guard of
		// their own.
		m1, m2 := mul1_128, mul2_128
		n := len(data) & 15
		var k1, k2 uint64
		if n != 0 {
			end := data[len(data)-16:]
			m := max(n, 8)
			k1 = load64(end[16-m:24-m]) >> (8 * uint(8-min(n, 8)))
			k2 = load64(end[8:]) >> (8 * uint(16-n))
		}
		// An index loop rather than a reslicing one: the compiler proves
		// the window from the loop bound, folds the index into the loads,
		// and has one counter to advance instead of a pointer, a length
		// and a capacity. The bound is hoisted by hand because Go 1.26
		// recomputes it every iteration otherwise; 1.27 hoists it itself.
		for i, end := 0, len(data)-16; i <= end; i += 16 {
			b := data[i : i+16]
			h1, h2 = mix128(h1, h2, load64(b), load64(b[8:]), m1, m2)
		}

		if n != 0 {
			k2 *= c2_128
			k2 = bits.RotateLeft64(k2, 33)
			k2 *= c1_128
			h2 ^= k2
			k1 *= c1_128
			k1 = bits.RotateLeft64(k1, 31)
			k1 *= c2_128
			h1 ^= k1
		}
	} else {
		// Every case knows its exact length, so the compiler proves all bounds
		// and each case is two or three word sized loads. This is faster than
		// the canonical byte at a time fallthrough switch.
		var k1, k2 uint64
		switch len(data) {
		case 15:
			k2 = uint64(load32(data[8:])) | uint64(load16(data[12:]))<<32 | uint64(data[14])<<48
			k1 = load64(data)
		case 14:
			k2 = uint64(load32(data[8:])) | uint64(load16(data[12:]))<<32
			k1 = load64(data)
		case 13:
			k2 = uint64(load32(data[8:])) | uint64(data[12])<<32
			k1 = load64(data)
		case 12:
			k2 = uint64(load32(data[8:]))
			k1 = load64(data)
		case 11:
			k2 = uint64(load16(data[8:])) | uint64(data[10])<<16
			k1 = load64(data)
		case 10:
			k2 = uint64(load16(data[8:]))
			k1 = load64(data)
		case 9:
			k2 = uint64(data[8])
			k1 = load64(data)
		case 8:
			k1 = load64(data)
		case 7:
			k1 = uint64(load32(data)) | uint64(load16(data[4:]))<<32 | uint64(data[6])<<48
		case 6:
			k1 = uint64(load32(data)) | uint64(load16(data[4:]))<<32
		case 5:
			k1 = uint64(load32(data)) | uint64(data[4])<<32
		case 4:
			k1 = uint64(load32(data))
		case 3:
			k1 = uint64(load16(data)) | uint64(data[2])<<16
		case 2:
			k1 = uint64(load16(data))
		case 1:
			k1 = uint64(data[0])
		}
		if len(data) > 8 {
			k2 *= c2_128
			k2 = bits.RotateLeft64(k2, 33)
			k2 *= c1_128
			h2 ^= k2
		}
		if len(data) > 0 {
			k1 *= c1_128
			k1 = bits.RotateLeft64(k1, 31)
			k1 *= c2_128
			h1 ^= k1
		}

	}

	h1 ^= uint64(clen)
	h2 ^= uint64(clen)

	h1 += h2
	h2 += h1

	h1 = fmix64(h1)
	h2 = fmix64(h2)

	h1 += h2
	h2 += h1

	return h1, h2
}
