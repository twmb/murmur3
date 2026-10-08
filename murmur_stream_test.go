package murmur3_test

import (
	"fmt"
	"testing"

	"github.com/twmb/murmur3"
)

func TestStreamingBoundaries(t *testing.T) {
	sizes := make([]int, 258)
	for i := range sizes {
		sizes[i] = i
	}
	sizes = append(sizes, 511, 512, 513, 16383, 16384, 16385, 262145)
	seeds := [][2]uint64{{0, 0}, {1, 7}, {1 << 63, 1<<63 + 1}, {^uint64(0), ^uint64(0) - 1}}
	for _, size := range sizes {
		for _, offset := range []int{0, 1, 7} {
			data := make([]byte, size+offset)[offset:]
			for i := range data {
				data[i] = byte(i*31 + i/251)
			}
			for _, seed := range seeds {
				want1, want2 := murmur3.SeedSum128(seed[0], seed[1], data)
				for _, step := range []int{1, 7, 15, 16, 17, 31, 32, 33, 63, 64, 65, 256} {
					h := murmur3.SeedNew128(seed[0], seed[1])
					for pos := 0; pos < len(data); pos += step {
						h.Write(nil)
						h.Write(data[pos:min(pos+step, len(data))])
					}
					if got1, got2 := h.Sum128(); got1 != want1 || got2 != want2 {
						t.Fatalf("size=%d offset=%d seed=%x step=%d: got %x/%x, want %x/%x", size, offset, seed, step, got1, got2, want1, want2)
					}

					h.Reset()
					h.Write(data)
					if got1, got2 := h.Sum128(); got1 != want1 || got2 != want2 {
						t.Fatalf("after reset size=%d offset=%d seed=%x step=%d: got %x/%x, want %x/%x", size, offset, seed, step, got1, got2, want1, want2)
					}
				}
			}
		}
	}
}

func BenchmarkStreaming64Small(b *testing.B) {
	for _, size := range []int{0, 16, 32, 64, 256, 1024} {
		b.Run(fmt.Sprintf("bytes%d", size), func(b *testing.B) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i*31 + i/251)
			}
			h := murmur3.New64()
			b.SetBytes(int64(size))
			b.ReportAllocs()
			for b.Loop() {
				h.Reset()
				h.Write(data)
				h.Sum64()
			}
		})
	}
}

func BenchmarkStreaming64Chunks(b *testing.B) {
	const segmentSize = 1 << 20
	split := []int{(16 << 10) - 25}
	for range 15 {
		split = append(split, 16<<10)
	}
	split = append(split, 25)
	for _, size := range []int{segmentSize, 8 * segmentSize} {
		for _, shape := range []struct {
			name  string
			sizes []int
		}{
			{"256KiB", []int{256 << 10}},
			{"16KiB", []int{16 << 10}},
			{"split", split},
		} {
			b.Run(fmt.Sprintf("object%dKiB/%s", size>>10, shape.name), func(b *testing.B) {
				var chunks [][]byte
				for pos, i := 0, 0; pos < size; i++ {
					chunk := make([]byte, min(shape.sizes[i%len(shape.sizes)], size-pos))
					for j := range chunk {
						k := pos + j
						chunk[j] = byte(k*31 + k/251)
					}
					chunks = append(chunks, chunk)
					pos += len(chunk)
				}
				b.SetBytes(int64(size))
				b.ReportAllocs()
				for b.Loop() {
					h := murmur3.New64()
					var sums []byte
					remaining := segmentSize
					for _, chunk := range chunks {
						for len(chunk) > 0 {
							n := min(len(chunk), remaining)
							h.Write(chunk[:n])
							chunk = chunk[n:]
							remaining -= n
							// Each MiB has an independent digest.
							if remaining == 0 {
								sums = h.Sum(sums)
								h.Reset()
								remaining = segmentSize
							}
						}
					}
				}
			})
		}
	}
}

func BenchmarkSum64MiB(b *testing.B) {
	data := make([]byte, 1<<20)
	for i := range data {
		data[i] = byte(i*31 + i/251)
	}
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	for b.Loop() {
		murmur3.Sum64(data)
	}
}
