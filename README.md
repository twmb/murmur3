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
under 64 bytes, for keys of varying length, and for every 32 bit sum, and 1
to 7 percent slower on fixed inputs of 64 bytes and more. The 32 bit assembly
was removed for the same reason back in Go 1.11. See the benchmarks below.

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
128Sizes/8192           4217    4479    4364     +3.5%   -2.6%
128Sizes/1024            534     576     557     +4.3%   -3.4%
128Sizes/256             140     159     149     +6.8%   -6.0%
128Sizes/64               48      52      49     +1.5%   -7.1%
128Sizes/32               35      36      34     -2.1%   -3.9%
128Branches/16            26      23      26     -1.4%  +10.8%
128Branches/13            25      34      21    -14.5%  -37.1%
128Branches/7             26      26      21    -18.5%  -19.7%
128Branches/3             25      22      21    -16.4%   -3.1%
64Sizes/8192            4216    4533    4362     +3.5%   -3.8%
64Sizes/32                35      38      34     -3.5%  -10.7%
32Sizes/8192           10216   10191    9585     -6.2%   -5.9%
32Sizes/64                89      88      76    -13.6%  -13.6%
32Sizes/32                50      49      41    -18.1%  -18.0%
32Branches/3              15      15      16     +5.6%   +5.8%
RandomLengths128         110      96      90    -18.1%   -5.8%
RandomLengths32           95      96      84    -11.6%  -12.0%
```

Per 16 byte block the assembly is 8.2 cycles and the Go loop 8.5, on both Go
1.26 and 1.27. The loop indexes data[i:i+16] with a hoisted bound rather than
reslicing, which is three fewer instructions per block. Holding the two
multipliers in registers instead of constants measured 8.1 on Go 1.27 and 9.3
on Go 1.26, the difference being one register name and its REX prefix, so
they stay constants. The tail is read out of the input's last 16 bytes with
two overlapping loads rather than a switch on its length, which is a jump
table and mispredicts on varying key lengths. The 32 bit sum sits on its four
cycle per block dependency chain.
