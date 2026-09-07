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
Go 1.27, the pure Go code is faster than the old assembly at every size
measured except 256 bytes, where it is 3 percent slower, and on Go 1.26 it is
8 percent slower from 256 bytes up. The 32 bit assembly was removed for the
same reason back in Go 1.11. See the benchmarks below.

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

Cycles per call from perf (`cycles:u` at a fixed iteration count, minimum of
five runs) on Go 1.27.1, Comet Lake i7-10710U, so CPU frequency and thermal
throttling drop out. `asm` is the amd64 assembly this library used to ship,
`old Go` the pure Go path other architectures ran, `new Go` the current code.
`Sizes` rows exercise the block loop, `Branches` rows the 0 to 16 byte tail,
`RandomLengths` rows draw an xorshift length in [0, 64) on every call.

```
benchmark                asm  old Go  new Go   new/asm new/old
128Sizes/8192           4215    4475    4134     -1.9%   -7.6%
128Sizes/1024            535     574     529     -1.1%   -7.9%
128Sizes/256             140     157     144     +2.6%   -8.3%
128Sizes/64               48      52      48     +0.6%   -7.9%
128Sizes/32               35      36      35     -1.1%   -3.2%
128Branches/16            26      24      26     -0.5%  +10.2%
128Branches/13            25      34      21    -16.5%  -38.8%
128Branches/7             26      26      21    -18.7%  -19.8%
128Branches/3             25      22      20    -20.3%   -7.3%
64Sizes/8192            4215    4533    4126     -2.1%   -9.0%
64Sizes/32                35      38      34     -2.7%   -9.5%
32Sizes/8192           10071   10050    9809     -2.6%   -2.4%
32Sizes/64                89      88      77    -13.6%  -13.5%
32Sizes/32                50      50      41    -17.7%  -17.6%
32Branches/3              15      15      16     +5.7%   +5.9%
RandomLengths128         110      96      88    -20.2%   -8.4%
RandomLengths32           96      96      85    -11.7%  -11.9%
```

Per 16 byte block the assembly is 8.2 cycles and the Go loop 8.1 on Go 1.27
and 9.3 on Go 1.26; the two compilers emit the same instructions with one
register named differently, and the REX prefix shifts the port binding onto
the multiply port. The tail is read out of the input's last 16 bytes with two
overlapping loads rather than a switch on its length, which is a jump table
and mispredicts on varying key lengths. The 32 bit sum sits on its four cycle
per block dependency chain.
