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
under 128 bytes and within a few percent above that. The 32 bit assembly was
removed for the same reason back in Go 1.11. See the benchmarks below.

Testing
=======

Testing includes comparing random inputs against the [canonical
implementation](https://github.com/aappleby/smhasher/blob/master/src/MurmurHash3.cpp),
and testing length 0 through 100 inputs to force all branches of the unrolled
loops and all tail lengths.

Because this code always reads input as little endian, testing against the
canonical source is skipped for big endian architectures. The canonical source
just converts bytes to numbers, meaning on big endian architectures, it will
use different numbers for its hashing.

Benchmarks
==========

Measured with perf on Go 1.27.1 amd64 (Comet Lake i7-10710U) as user space
cycles per call at a fixed iteration count, so CPU frequency and thermal
throttling drop out. `asm` is the hand rolled amd64 assembly this library
used to ship, `old Go` is the pure Go path every other architecture ran
before, and `new Go` is the current code. The `Sizes` rows hash the block
loop; the `Branches` rows hash 0 to 16 bytes and exercise the tail. Each row
includes a few cycles of benchmark harness.

```
benchmark            asm  old Go  new Go   new/asm new/old
128Sizes/8192       4187    4450    4662    +11.4%   +4.8%
128Sizes/1024        530     570     619    +16.8%   +8.6%
128Sizes/256         139     158     160    +14.8%   +1.0%
128Sizes/64           48      52      51     +7.3%   -1.8%
128Sizes/32           34      35      34     -1.1%   -2.8%
128Branches/16        26      23      26     -0.9%  +11.0%
128Branches/13        25      34      22    -12.5%  -35.8%
128Branches/7         26      26      22    -15.3%  -17.7%
128Branches/3         25      21      21    -16.6%   -2.8%
64Sizes/8192        4184    4502    4663    +11.4%   +3.6%
32Sizes/8192       10030   10008    9771     -2.6%   -2.4%
32Sizes/64            87      87      79     -9.8%   -9.7%
32Branches/3          15      15      17    +13.3%  +13.3%
```

Per 16 byte block that is 8.2 cycles for the assembly and about 9.2 for the
Go loop. The loop is bound by latency and by port pressure, not by
instruction count: the four 64 bit multiplies per block and the LEAs share
one execution port, and the loop carried h1 to h2 chain runs through those
same instructions. The compiler already does the right thing here. It folds
5*c1 + c2 into a single constant, which takes an add off that chain, and
every attempt to help it (a four block unroll, constants hoisted into
registers, a strictly greater loop bound to drop the reslice guard) produced
fewer instructions and more cycles, up to 11.3 per block. What the assembly
still buys is 22 instructions per block against 27 with the chain and port
already saturated. Byte identical loops also land anywhere from 8.8 to 11
cycles per block depending on which registers the allocator hands out,
because a three address add becomes an LEA on the contended port, so a few
percent either way between two Go versions is noise you cannot steer from
source.

The 32 bit sum is bound by its own four cycle per block dependency chain and
sits at about five; nothing above the algorithm changes that. Its gain at
small sizes comes from the two block unroll.
