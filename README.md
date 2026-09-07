murmur3
=======

[![Go Reference](https://pkg.go.dev/badge/github.com/twmb/murmur3.svg)](https://pkg.go.dev/github.com/twmb/murmur3)
[![ci](https://github.com/twmb/murmur3/actions/workflows/ci.yml/badge.svg)](https://github.com/twmb/murmur3/actions/workflows/ci.yml)

Native Go implementation of Austin Appleby's third MurmurHash revision (aka
MurmurHash3).

Includes 32, 64, and 128 bit sums, seeding functions, string functions that
hash without converting to a slice, and streaming hashes implementing Go's
standard [Hash](https://pkg.go.dev/hash#Hash) and
[Cloner](https://pkg.go.dev/hash#Cloner) interfaces.

This library started as a fork of [spaolacci/murmur3](https://github.com/spaolacci/murmur3).
The reference algorithm has been slightly hacked as to support the streaming
mode required by Go's standard Hash interface.

Endianness
==========

Unlike the canonical source, this library **always** reads bytes as little
endian numbers. This makes the hashes portable across architectures, although
does mean that hashing is a bit slower on big endian architectures.

Safety
======

This library uses no `unsafe`. Bytes are read with plain indexing, which the
compiler turns into single word sized loads on architectures that allow
unaligned access, and the loops are shaped so that the compiler proves every
index in bounds. Strings are hashed through one implementation generic over
`string | []byte`, so hashing a string does not copy it and does not
allocate.

Assembly
========

Earlier versions shipped hand rolled amd64 assembly for the 64 and 128 bit
sums. That assembly was removed once the compiler's output caught up: as of
Go 1.27, the pure Go code is faster than the old assembly for every input
under 32 bytes and within 3 to 8 percent above that. The 32 bit assembly was
removed for the same reason back in Go 1.11. See the benchmarks below.

Testing
=======

Testing includes comparing random inputs against the [canonical
implementation](https://github.com/aappleby/smhasher/blob/master/src/MurmurHash3.cpp),
and testing length 0 through 100 inputs to force the block loop, the trailing
block, and all tail lengths.

Because this code always reads input as little endian, testing against the
canonical source is skipped for big endian architectures. The canonical source
just converts bytes to numbers, meaning on big endian architectures, it will
use different numbers for its hashing.

Benchmarks
==========

Measured with perf on Go 1.27.1 amd64 (Comet Lake i7-10710U) as user space
cycles per call at a fixed iteration count, minimum of five runs, so CPU
frequency and thermal throttling drop out. `asm` is the hand rolled amd64
assembly this library used to ship, `old Go` is the pure Go path every other
architecture ran before, and `new Go` is the current code. The `Sizes` rows
hash the block loop; the `Branches` rows hash 0 to 16 bytes and exercise the
tail. Each row includes a few cycles of benchmark harness.

```
benchmark            asm  old Go  new Go   new/asm new/old
128Sizes/8192       4212    4477    4354     +3.4%   -2.7%
128Sizes/1024        534     574     558     +4.6%   -2.7%
128Sizes/256         140     156     151     +7.9%   -3.6%
128Sizes/64           48      52      50     +4.4%   -4.3%
128Sizes/32           35      36      35     -0.8%   -2.8%
128Branches/16        26      23      26     -1.3%  +10.8%
128Branches/13        25      34      22    -12.6%  -35.6%
128Branches/7         26      26      22    -14.4%  -15.9%
128Branches/3         25      22      21    -16.5%   -2.9%
64Sizes/8192        4217    4532    4357     +3.3%   -3.9%
32Sizes/8192       10076   10048    9765     -3.1%   -2.8%
32Sizes/64            89      89      80    -10.1%  -10.1%
32Sizes/32            50      50      44    -10.9%  -10.8%
32Branches/3          15      15      17    +12.3%  +12.4%
```

Per 16 byte block that is 8.2 cycles for the assembly and 8.5 for the Go
loop, and perf's port counters say where the difference is. The Go loop
issues 29 micro ops per block to the assembly's 19, so on a four wide core its
floor is 7.3 cycles and ports 0, 1 and 6 each run 75 to 80 percent busy; the
measured 8.5 is scheduling slack on top of that. The assembly has slack on
every port and sits on its dependency chain instead: the four 64 bit
multiplies and the rotates are off that chain, and what is left is h1 to h2
through two three operand LEAs at three cycles each.

The one change that helped was the strictly greater loop bound, which
removes the zero length pointer guard the compiler emits after every reslice,
four micro ops per block, without touching the arithmetic. Everything else
was measured and lost: unrolling two or four blocks per iteration, hoisting
the constants into registers, folding the two adds by hand, and moving the
loop into its own function. Cutting micro ops only wins if the h1 to h2 chain
stays at the five cycles the compiler's own folding of 5*c1 + c2 gives it, and
which registers the allocator hands out decides whether a three address add
is an ADD or an LEA. Byte identical loops landed anywhere from 8.5 to 11
cycles per block on that alone. That also makes the result a property of the
compiler version: Go 1.26 and 1.27 both produce the 8.5 cycle loop from this
source, while Go 1.25 emits four more instructions for it and lands at 9.5,
slower than the two block unroll it replaced. GOAMD64=v3 changes nothing.

The 32 bit sum is bound by its own four cycle per block dependency chain and
sits at 4.8 with every port under 80 percent busy; nothing above the
algorithm changes that. The strict bound
is what it gains at 32 to 64 bytes.
