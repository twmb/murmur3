package testdata

// #cgo CFLAGS: -std=gnu99
// #include <stdint.h>
// #include "MurmurHash3.cpp"
// #include "MurmurHash3.h"
import "C"

import "unsafe"

func SeedSum32(seed uint32, data []byte) uint32 {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	var out uint32
	C.MurmurHash3_x86_32(p, C.int(len(data)), C.uint32_t(seed), unsafe.Pointer(&out))
	return out
}

func SeedSum64(seed uint32, data []byte) uint64 {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	var out struct {
		h1 uint64
		h2 uint64
	}
	C.MurmurHash3_x64_128(p, C.int(len(data)), C.uint32_t(seed), unsafe.Pointer(&out))
	return out.h1
}

func SeedSum128(seed uint32, data []byte) (h1, h2 uint64) {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	var out struct {
		h1 uint64
		h2 uint64
	}
	C.MurmurHash3_x64_128(p, C.int(len(data)), C.uint32_t(seed), unsafe.Pointer(&out))
	return out.h1, out.h2
}
