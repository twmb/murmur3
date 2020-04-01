package murmur3

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"strconv"
	"testing"
	"testing/quick"
	"unsafe"

	"github.com/twmb/murmur3/testdata"
)

var isLittleEndian = func() bool {
	i := uint16(1)
	return (*(*[2]byte)(unsafe.Pointer(&i)))[0] == 1
}()

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

func TestQuickSum32(t *testing.T) {
	f := func(data []byte) bool {
		goh1 := Sum32(data)
		goh2 := StringSum32(string(data))
		cpph1 := goh1
		if isLittleEndian {
			cpph1 = testdata.SeedSum32(0, data)
		}
		return goh1 == goh2 && goh1 == cpph1
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestQuickSeedSum32(t *testing.T) {
	f := func(seed uint32, data []byte) bool {
		goh1 := SeedSum32(seed, data)
		goh2 := SeedStringSum32(seed, string(data))
		goh3 := func() uint32 { h := SeedNew32(seed); h.Write(data); return binary.BigEndian.Uint32(h.Sum(nil)) }()
		cpph1 := goh1
		if isLittleEndian {
			cpph1 = testdata.SeedSum32(seed, data)
		}
		return goh1 == goh2 && goh1 == goh3 && goh1 == cpph1
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestQuickSum64(t *testing.T) {
	f := func(data []byte) bool {
		goh1 := Sum64(data)
		goh2 := StringSum64(string(data))
		cpph1 := goh1
		if isLittleEndian {
			cpph1 = testdata.SeedSum64(0, data)
		}
		return goh1 == goh2 && goh1 == cpph1
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestQuickSeedSum64(t *testing.T) {
	f := func(seed uint32, data []byte) bool {
		goh1 := SeedSum64(uint64(seed), data)
		goh2 := SeedStringSum64(uint64(seed), string(data))
		goh3 := func() uint64 { h := SeedNew64(uint64(seed)); h.Write(data); return binary.BigEndian.Uint64(h.Sum(nil)) }()
		cpph1 := goh1
		if isLittleEndian {
			cpph1 = testdata.SeedSum64(seed, data)
		}
		return goh1 == goh2 && goh1 == goh3 && goh1 == cpph1
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestQuickSum128(t *testing.T) {
	f := func(data []byte) bool {
		goh1, goh2 := Sum128(data)
		goh3, goh4 := StringSum128(string(data))
		cpph1, cpph2 := goh1, goh2
		if isLittleEndian {
			cpph1, cpph2 = testdata.SeedSum128(0, data)
		}
		return goh1 == goh3 && goh2 == goh4 && goh1 == cpph1 && goh2 == cpph2
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestQuickSeedSum128(t *testing.T) {
	f := func(seed uint32, data []byte) bool {
		goh1, goh2 := SeedSum128(uint64(seed), uint64(seed), data)
		goh3, goh4 := SeedStringSum128(uint64(seed), uint64(seed), string(data))
		goh5, goh6 := func() (uint64, uint64) {
			h := SeedNew128(uint64(seed), uint64(seed))
			h.Write(data)
			sum := h.Sum(nil)
			return binary.BigEndian.Uint64(sum), binary.BigEndian.Uint64(sum[8:])
		}()
		cpph1, cpph2 := goh1, goh2
		if isLittleEndian {
			testdata.SeedSum128(seed, data)
		}
		return goh1 == goh3 && goh2 == goh4 &&
			goh1 == goh5 && goh2 == goh6 &&
			goh1 == cpph1 && goh2 == cpph2
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// go1.14 showed that doing *(*uint32)(unsafe.Pointer(&data[i*4])) was unsafe
// due to alignment issues; this test ensures that we will always catch that.
func TestUnaligned(t *testing.T) {
	in1 := []byte("abcdefghijklmnopqrstuvwxyz")
	in2 := []byte("_abcdefghijklmnopqrstuvwxyz")

	{
		sum1 := Sum32(in1)
		sum2 := Sum32(in2[1:])
		if sum1 != sum2 {
			t.Errorf("%s: got sum1 %v sum2 %v unexpectedly not equal", "Sum32", sum1, sum2)
		}
	}

	{
		sum1 := Sum64(in1)
		sum2 := Sum64(in2[1:])
		if sum1 != sum2 {
			t.Errorf("%s: got sum1 %v sum2 %v unexpectedly not equal", "Sum64", sum1, sum2)
		}
	}

	{
		sum1l, sum1r := Sum128(in1)
		sum2l, sum2r := Sum128(in2[1:])
		if sum1l != sum2l {
			t.Errorf("%s: got sum1l %v sum2l %v unexpectedly not equal", "Sum128 left", sum1l, sum2l)
		}
		if sum1r != sum2r {
			t.Errorf("%s: got sum1r %v sum2r %v unexpectedly not equal", "Sum128 right", sum1r, sum2r)
		}
	}

	{
		sum1 := func() uint32 { n := New32(); n.Write(in1); return n.Sum32() }()
		sum2 := func() uint32 { n := New32(); n.Write(in2[1:]); return n.Sum32() }()
		if sum1 != sum2 {
			t.Errorf("%s: got sum1 %v sum2 %v unexpectedly not equal", "New32", sum1, sum2)
		}
	}

	{
		sum1 := func() uint64 { n := New64(); n.Write(in1); return n.Sum64() }()
		sum2 := func() uint64 { n := New64(); n.Write(in2[1:]); return n.Sum64() }()
		if sum1 != sum2 {
			t.Errorf("%s: got sum1 %v sum2 %v unexpectedly not equal", "New64", sum1, sum2)
		}
	}

}

// TestBoundaries forces every block/tail path to be exercised for Sum32 and Sum128.
func TestBoundaries(t *testing.T) {
	const maxCheck = 17
	var data [maxCheck]byte
	for i := 0; !t.Failed() && i < 20; i++ {
		// Check all zeros the first iteration.
		for size := 0; size <= maxCheck; size++ {
			test := data[:size]
			g32h1 := Sum32(test)
			g32h1s := SeedSum32(0, test)
			c32h1 := g32h1
			if isLittleEndian {
				c32h1 = testdata.SeedSum32(0, test)
			}
			if g32h1 != c32h1 {
				t.Errorf("size #%d: in: %x, g32h1 (%d) != c32h1 (%d); attempt #%d", size, test, g32h1, c32h1, i)
			}
			if g32h1s != c32h1 {
				t.Errorf("size #%d: in: %x, gh32h1s (%d) != c32h1 (%d); attempt #%d", size, test, g32h1s, c32h1, i)
			}
			g64h1 := Sum64(test)
			g64h1s := SeedSum64(0, test)
			c64h1 := g64h1
			if isLittleEndian {
				c64h1 = testdata.SeedSum64(0, test)
			}
			if g64h1 != c64h1 {
				t.Errorf("size #%d: in: %x, g64h1 (%d) != c64h1 (%d); attempt #%d", size, test, g64h1, c64h1, i)
			}
			if g64h1s != c64h1 {
				t.Errorf("size #%d: in: %x, g64h1s (%d) != c64h1 (%d); attempt #%d", size, test, g64h1s, c64h1, i)
			}
			g128h1, g128h2 := Sum128(test)
			g128h1s, g128h2s := SeedSum128(0, 0, test)
			c128h1, c128h2 := g128h1, g128h2
			if isLittleEndian {
				c128h1, c128h2 = testdata.SeedSum128(0, test)
			}
			if g128h1 != c128h1 {
				t.Errorf("size #%d: in: %x, g128h1 (%d) != c128h1 (%d); attempt #%d", size, test, g128h1, c128h1, i)
			}
			if g128h2 != c128h2 {
				t.Errorf("size #%d: in: %x, g128h2 (%d) != c128h2 (%d); attempt #%d", size, test, g128h2, c128h2, i)
			}
			if g128h1s != c128h1 {
				t.Errorf("size #%d: in: %x, g128h1s (%d) != c128h1 (%d); attempt #%d", size, test, g128h1s, c128h1, i)
			}
			if g128h2s != c128h2 {
				t.Errorf("size #%d: in: %x, g128h2s (%d) != c128h2 (%d); attempt #%d", size, test, g128h2s, c128h2, i)
			}
		}
		// Randomize the data for all subsequent tests.
		io.ReadFull(rand.Reader, data[:])
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

// Our lengths force 1) the function base itself (no loop/tail), 2) remainders
// and 3) the loop itself.

func Benchmark32Branches(b *testing.B) {
	for length := 0; length <= 4; length++ {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf := make([]byte, length)
			b.SetBytes(int64(length))
			b.ResetTimer()
			for length := 0; length < b.N; length++ {
				Sum32(buf)
			}
		})
	}
}

func BenchmarkPartial32Branches(b *testing.B) {
	for length := 0; length <= 4; length++ {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf := make([]byte, length)
			b.SetBytes(int64(length))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hasher := New32()
				hasher.Write(buf)
				hasher.Sum32()
			}
		})
	}
}

func Benchmark128Branches(b *testing.B) {
	for length := 0; length <= 16; length++ {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf := make([]byte, length)
			b.SetBytes(int64(length))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Sum128(buf)
			}
		})
	}
}

// Sizes below pick up where branches left off to demonstrate speed at larger
// slice sizes.

func Benchmark32Sizes(b *testing.B) {
	buf := make([]byte, 8192)
	for length := 32; length <= cap(buf); length *= 2 {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf = buf[:length]
			b.SetBytes(int64(length))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Sum32(buf)
			}
		})
	}
}

func Benchmark64Sizes(b *testing.B) {
	buf := make([]byte, 8192)
	for length := 32; length <= cap(buf); length *= 2 {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf = buf[:length]
			b.SetBytes(int64(length))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Sum64(buf)
			}
		})
	}
}

func Benchmark128Sizes(b *testing.B) {
	buf := make([]byte, 8192)
	for length := 32; length <= cap(buf); length *= 2 {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf = buf[:length]
			b.SetBytes(int64(length))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Sum128(buf)
			}
		})
	}
}

func BenchmarkNoescape32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf [8192]byte
		Sum32(buf[:])
	}
}

func BenchmarkNoescape128(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf [8192]byte
		Sum128(buf[:])
	}
}

func BenchmarkStrslice(b *testing.B) {
	var s []byte
	for i := 0; i < b.N; i++ {
		strslice(s)
	}
}
