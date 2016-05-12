package murmur3

import (
	"crypto/rand"
	"fmt"
	"hash"
	"io"
	"testing"
	"testing/quick"

	"github.com/twmb/murmur3/testdata"
)

var data = []struct {
	h32   uint32
	h64_1 uint64
	h64_2 uint64
	s     string
}{
	{0x00000000, 0x0000000000000000, 0x0000000000000000, ""},
	{0x248bfa47, 0xcbd8a7b341bd9b02, 0x5b1e906a48ae1d19, "hello"},
	{0x149bbb7f, 0x342fac623a5ebc8e, 0x4cdcbc079642414d, "hello, world"},
	{0xe31e8a70, 0xb89e5988b737affc, 0x664fc2950231b2cb, "19 Jan 2038 at 3:14:07 AM"},
	{0xd5c48bfc, 0xcd99481f9ee902c9, 0x695da1a38987b6e7, "The quick brown fox jumps over the lazy dog."},
}

func TestRef(t *testing.T) {
	for _, elem := range data {

		var h32 hash.Hash32 = New32()
		h32.Write([]byte(elem.s))
		if v := h32.Sum32(); v != elem.h32 {
			t.Errorf("'%s': 0x%x (want 0x%x)", elem.s, v, elem.h32)
		}

		var h32_byte hash.Hash32 = New32()
		h32_byte.Write([]byte(elem.s))
		target := fmt.Sprintf("%08x", elem.h32)
		if p := fmt.Sprintf("%x", h32_byte.Sum(nil)); p != target {
			t.Errorf("'%s': %s (want %s)", elem.s, p, target)
		}

		if v := Sum32([]byte(elem.s)); v != elem.h32 {
			t.Errorf("'%s': 0x%x (want 0x%x)", elem.s, v, elem.h32)
		}

		var h64 hash.Hash64 = New64()
		h64.Write([]byte(elem.s))
		if v := h64.Sum64(); v != elem.h64_1 {
			t.Errorf("'%s': 0x%x (want 0x%x)", elem.s, v, elem.h64_1)
		}

		var h64_byte hash.Hash64 = New64()
		h64_byte.Write([]byte(elem.s))
		target = fmt.Sprintf("%016x", elem.h64_1)
		if p := fmt.Sprintf("%x", h64_byte.Sum(nil)); p != target {
			t.Errorf("Sum64: '%s': %s (want %s)", elem.s, p, target)
		}

		if v := Sum64([]byte(elem.s)); v != elem.h64_1 {
			t.Errorf("Sum64: '%s': 0x%x (want 0x%x)", elem.s, v, elem.h64_1)
		}

		var h128 Hash128 = New128()
		h128.Write([]byte(elem.s))
		if v1, v2 := h128.Sum128(); v1 != elem.h64_1 || v2 != elem.h64_2 {
			t.Errorf("New128: '%s': 0x%x-0x%x (want 0x%x-0x%x)", elem.s, v1, v2, elem.h64_1, elem.h64_2)
		}

		var h128_byte Hash128 = New128()
		h128_byte.Write([]byte(elem.s))
		target = fmt.Sprintf("%016x%016x", elem.h64_1, elem.h64_2)
		if p := fmt.Sprintf("%x", h128_byte.Sum(nil)); p != target {
			t.Errorf("New128: '%s': %s (want %s)", elem.s, p, target)
		}

		if v1, v2 := Sum128([]byte(elem.s)); v1 != elem.h64_1 || v2 != elem.h64_2 {
			t.Errorf("Sum128: '%s': 0x%x-0x%x (want 0x%x-0x%x)", elem.s, v1, v2, elem.h64_1, elem.h64_2)
		}
	}
}

func TestQuickSeedSum32(t *testing.T) {
	f := func(seed uint32, data []byte) bool {
		goh1 := SeedSum32(seed, data)
		cpph1 := testdata.SeedSum32(seed, data)
		return goh1 == cpph1
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestQuickSeedSum64(t *testing.T) {
	f := func(seed uint32, data []byte) bool {
		goh1 := SeedSum64(uint64(seed), data)
		cpph1 := testdata.SeedSum64(seed, data)
		return goh1 == cpph1
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestQuickSeedSum128(t *testing.T) {
	f := func(seed uint32, data []byte) bool {
		goh1, goh2 := SeedSum128(uint64(seed), uint64(seed), data)
		cpph1, cpph2 := testdata.SeedSum128(seed, data)
		return goh1 == cpph1 && goh2 == cpph2
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestBoundaries forces every block/tail path to be exercised for Sum32 and Sum128.
func TestBoundaries(t *testing.T) {
	var data [17]byte
	for i := 0; !t.Failed() && i < 20; i++ {
		io.ReadFull(rand.Reader, data[:])
		for size := 0; size <= 17; size++ {
			test := data[:size]
			g32h1 := Sum32(test)
			c32h1 := testdata.SeedSum32(0, test)
			if g32h1 != c32h1 {
				t.Errorf("size #%d: in: %x, g32h1 (%d) != c32h1 (%d); attempt #%d", size, test, g32h1, c32h1, i)
			}
			g64h1 := Sum64(test)
			c64h1 := testdata.SeedSum64(0, test)
			if g64h1 != c64h1 {
				t.Errorf("size #%d: in: %x, g64h1 (%d) != c64h1 (%d); attempt #%d", size, test, g64h1, c64h1, i)
			}
			g128h1, g128h2 := Sum128(test)
			c128h1, c128h2 := testdata.SeedSum128(0, test)
			if g128h1 != c128h1 {
				t.Errorf("size #%d: in: %x, g128h1 (%d) != c128h1 (%d); attempt #%d", size, test, g128h1, c128h1, i)
			}
			if g128h2 != c128h2 {
				t.Errorf("size #%d: in: %x, g128h2 (%d) != c128h2 (%d); attempt #%d", size, test, g128h2, c128h2, i)
			}
		}
	}
}

func TestIncremental(t *testing.T) {
	for _, elem := range data {
		h32 := New32()
		h128 := New128()
		for i, j, k := 0, 0, len(elem.s); i < k; i = j {
			j = 2*i + 3
			if j > k {
				j = k
			}
			s := elem.s[i:j]
			print(s + "|")
			h32.Write([]byte(s))
			h128.Write([]byte(s))
		}
		println()
		if v := h32.Sum32(); v != elem.h32 {
			t.Errorf("'%s': 0x%x (want 0x%x)", elem.s, v, elem.h32)
		}
		if v1, v2 := h128.Sum128(); v1 != elem.h64_1 || v2 != elem.h64_2 {
			t.Errorf("'%s': 0x%x-0x%x (want 0x%x-0x%x)", elem.s, v1, v2, elem.h64_1, elem.h64_2)
		}
	}
}

//---

func bench32(b *testing.B, length int) {
	buf := make([]byte, length)
	b.SetBytes(int64(length))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sum32(buf)
	}
}

func BenchmarkBase32(b *testing.B) {
	bench32(b, 0)
}
func BenchmarkLoop32(b *testing.B) {
	bench32(b, 4)
}
func BenchmarkTail32_1(b *testing.B) {
	bench32(b, 1)
}
func BenchmarkTail32_2(b *testing.B) {
	bench32(b, 2)
}
func BenchmarkTail32_3(b *testing.B) {
	bench32(b, 3)
}

//---

func benchPartial32(b *testing.B, length int) {
	buf := make([]byte, length)
	b.SetBytes(int64(length))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher := New32()
		hasher.Write(buf)
		hasher.Sum32()
	}
}

func BenchmarkBasePartial32(b *testing.B) {
	benchPartial32(b, 0)
}
func BenchmarkLoopPartial32(b *testing.B) {
	benchPartial32(b, 4)
}
func BenchmarkTailPartial32_1(b *testing.B) {
	benchPartial32(b, 1)
}
func BenchmarkTailPartial32_2(b *testing.B) {
	benchPartial32(b, 2)
}
func BenchmarkTailPartial32_3(b *testing.B) {
	benchPartial32(b, 3)
}

//---

func bench128(b *testing.B, length int) {
	buf := make([]byte, length)
	b.SetBytes(int64(length))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sum128(buf)
	}
}

func BenchmarkBase128(b *testing.B) {
	bench128(b, 0)
}
func BenchmarkLoop128(b *testing.B) {
	bench128(b, 16)
}
func BenchmarkTail128_1(b *testing.B) {
	bench128(b, 1)
}
func BenchmarkTail128_2(b *testing.B) {
	bench128(b, 2)
}
func BenchmarkTail128_3(b *testing.B) {
	bench128(b, 3)
}
func BenchmarkTail128_4(b *testing.B) {
	bench128(b, 4)
}
func BenchmarkTail128_5(b *testing.B) {
	bench128(b, 5)
}
func BenchmarkTail128_6(b *testing.B) {
	bench128(b, 6)
}
func BenchmarkTail128_7(b *testing.B) {
	bench128(b, 7)
}
func BenchmarkTail128_8(b *testing.B) {
	bench128(b, 8)
}
func BenchmarkTail128_9(b *testing.B) {
	bench128(b, 9)
}
func BenchmarkTail128_10(b *testing.B) {
	bench128(b, 10)
}
func BenchmarkTail128_11(b *testing.B) {
	bench128(b, 11)
}
func BenchmarkTail128_12(b *testing.B) {
	bench128(b, 12)
}
func BenchmarkTail128_13(b *testing.B) {
	bench128(b, 13)
}
func BenchmarkTail128_14(b *testing.B) {
	bench128(b, 14)
}
func BenchmarkTail128_15(b *testing.B) {
	bench128(b, 15)
}

//---
