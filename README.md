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

All numbers below are from Go 1.27.1 on an amd64 laptop (i7-10710U) that
thermally throttles, so treat them as directional: the three builds were run
interleaved, round robin, pinned to one core, nine rounds each, and benchstat
reports the median and spread. On a machine that does not throttle the spread
should collapse to a percent or two. To reproduce, build the test binary at
the old and new commits and alternate runs of each.

The `Branches` benchmarks hash 0 to 16 bytes and exercise the tail switch.
The `Sizes` benchmarks hash 32 bytes to 8 KiB and exercise the block loop.

Old hand rolled amd64 assembly (left) versus the current pure Go (right):

```
               │ rr_asm_c.txt │             rr_new_c.txt             │
               │    sec/op    │    sec/op      vs base               │
32Branches/0     2.635n ±  7%    2.726n ± 10%   +3.45% (p=0.008 n=9)
32Branches/1     2.653n ± 12%    2.713n ± 13%        ~ (p=0.258 n=9)
32Branches/2     2.925n ± 11%    3.263n ±  8%  +11.56% (p=0.014 n=9)
32Branches/3     3.207n ±  7%    3.435n ± 10%   +7.11% (p=0.019 n=9)
32Branches/4     3.334n ± 11%    3.415n ± 16%        ~ (p=0.621 n=9)
128Branches/0    3.629n ± 12%    3.813n ± 11%   +5.07% (p=0.024 n=9)
128Branches/1    4.742n ± 11%    4.113n ± 10%  -13.26% (p=0.001 n=9)
128Branches/2    5.147n ± 12%    4.192n ± 10%  -18.55% (p=0.000 n=9)
128Branches/3    5.375n ± 11%    4.367n ± 10%  -18.75% (p=0.000 n=9)
128Branches/4    4.577n ± 10%    4.012n ± 13%  -12.34% (p=0.001 n=9)
128Branches/5    4.977n ± 12%    4.233n ± 10%  -14.95% (p=0.000 n=9)
128Branches/6    5.271n ±  9%    4.293n ±  8%  -18.55% (p=0.000 n=9)
128Branches/7    5.568n ± 10%    4.461n ± 10%  -19.88% (p=0.000 n=9)
128Branches/8    4.463n ± 11%    4.098n ±  7%   -8.18% (p=0.002 n=9)
128Branches/9    5.230n ±  9%    4.446n ±  5%  -14.99% (p=0.000 n=9)
128Branches/10   5.683n ±  8%    4.223n ± 11%  -25.69% (p=0.000 n=9)
128Branches/11   6.021n ±  7%    4.573n ±  7%  -24.05% (p=0.000 n=9)
128Branches/12   4.937n ± 12%    4.249n ±  9%  -13.94% (p=0.000 n=9)
128Branches/13   5.398n ± 10%    4.482n ±  9%  -16.97% (p=0.000 n=9)
128Branches/14   5.874n ± 10%    4.473n ±  9%  -23.85% (p=0.000 n=9)
128Branches/15   5.726n ± 11%    4.714n ±  8%  -17.67% (p=0.000 n=9)
128Branches/16   5.642n ± 11%    5.388n ± 13%        ~ (p=0.074 n=9)
32Sizes/32       10.80n ± 10%    10.12n ±  7%   -6.30% (p=0.011 n=9)
32Sizes/64       19.31n ± 10%    17.23n ± 12%  -10.77% (p=0.006 n=9)
32Sizes/128      37.02n ± 10%    34.64n ±  8%   -6.43% (p=0.011 n=9)
32Sizes/256      72.01n ±  9%    74.52n ±  8%        ~ (p=0.436 n=9)
32Sizes/512      145.9n ±  9%    137.5n ± 10%   -5.76% (p=0.014 n=9)
32Sizes/1024     291.2n ±  7%    279.1n ± 11%        ~ (p=0.077 n=9)
32Sizes/2048     569.6n ± 12%    558.6n ± 11%        ~ (p=0.161 n=9)
32Sizes/4096     1.148µ ±  8%    1.117µ ± 10%        ~ (p=0.214 n=9)
32Sizes/8192     2.227µ ± 10%    2.206µ ± 11%        ~ (p=0.297 n=9)
64Sizes/32       7.591n ±  7%    7.907n ± 13%        ~ (p=0.077 n=9)
64Sizes/64       10.68n ±  7%    11.72n ±  6%   +9.74% (p=0.001 n=9)
64Sizes/128      17.20n ± 10%    19.96n ±  6%  +16.05% (p=0.000 n=9)
64Sizes/256      31.86n ±  6%    36.51n ±  2%  +14.60% (p=0.000 n=9)
64Sizes/512      60.02n ±  8%    68.54n ±  4%  +14.20% (p=0.000 n=9)
64Sizes/1024     118.8n ±  7%    140.8n ±  2%  +18.52% (p=0.000 n=9)
64Sizes/2048     242.1n ±  6%    263.4n ±  2%   +8.80% (p=0.000 n=9)
64Sizes/4096     468.4n ±  8%    534.7n ±  2%  +14.15% (p=0.000 n=9)
64Sizes/8192     945.7n ±  6%   1058.0n ±  2%  +11.87% (p=0.000 n=9)
128Sizes/32      7.494n ±  9%    7.267n ±  4%        ~ (p=0.161 n=9)
128Sizes/64      10.31n ± 10%    11.09n ±  5%   +7.57% (p=0.038 n=9)
128Sizes/128     17.00n ±  9%    19.46n ±  4%  +14.47% (p=0.000 n=9)
128Sizes/256     30.81n ±  6%    36.00n ±  7%  +16.85% (p=0.000 n=9)
128Sizes/512     60.98n ±  4%    69.94n ±  5%  +14.69% (p=0.000 n=9)
128Sizes/1024    117.2n ±  8%    140.9n ±  8%  +20.22% (p=0.000 n=9)
128Sizes/2048    236.1n ±  8%    271.7n ±  4%  +15.08% (p=0.000 n=9)
128Sizes/4096    467.7n ±  8%    547.3n ±  5%  +17.02% (p=0.000 n=9)
128Sizes/8192    935.5n ±  5%   1065.0n ±  9%  +13.84% (p=0.000 n=9)
geomean          23.29n          22.88n         -1.77%

               │ rr_asm_c.txt  │             rr_new_c.txt              │
```

The pure Go code is faster than the assembly was for every 128 bit input
under 64 bytes, in particular the 9 through 15 byte tails, and the 32 bit
sums are faster at every size from 32 bytes up. The block loop for the 64 and
128 bit sums remains 10 to 20 percent slower than the assembly on inputs of
128 bytes and more. That is the slice bookkeeping the compiler emits per
iteration; on this machine the two block unroll did not measurably reduce it,
and against the old plain loop it is within noise below 512 bytes and a few
percent behind above.

Old pure Go (left, what every non-amd64 architecture ran) versus the current
pure Go (right), 64 and 128 bit sums:

```
               │ rr_gen_c.txt  │            rr_new_c.txt             │
               │    sec/op     │    sec/op     vs base               │
128Branches/0     3.326n ± 10%   3.813n ± 11%  +14.64% (p=0.000 n=9)
128Branches/1     4.021n ±  9%   4.113n ± 10%        ~ (p=0.297 n=9)
128Branches/2     4.402n ±  7%   4.192n ± 10%        ~ (p=0.161 n=9)
128Branches/3     4.615n ± 11%   4.367n ± 10%        ~ (p=0.161 n=9)
128Branches/4     4.815n ± 13%   4.012n ± 13%  -16.68% (p=0.000 n=9)
128Branches/5     5.126n ± 11%   4.233n ± 10%  -17.42% (p=0.000 n=9)
128Branches/6     5.350n ±  9%   4.293n ±  8%  -19.76% (p=0.000 n=9)
128Branches/7     5.616n ±  9%   4.461n ± 10%  -20.57% (p=0.000 n=9)
128Branches/8     5.540n ± 10%   4.098n ±  7%  -26.03% (p=0.000 n=9)
128Branches/9     6.000n ± 10%   4.446n ±  5%  -25.90% (p=0.000 n=9)
128Branches/10    6.456n ± 10%   4.223n ± 11%  -34.59% (p=0.000 n=9)
128Branches/11    6.781n ± 11%   4.573n ±  7%  -32.56% (p=0.000 n=9)
128Branches/12    7.290n ± 11%   4.249n ±  9%  -41.71% (p=0.000 n=9)
128Branches/13    7.499n ± 14%   4.482n ±  9%  -40.23% (p=0.000 n=9)
128Branches/14    8.063n ± 21%   4.473n ±  9%  -44.52% (p=0.000 n=9)
128Branches/15    8.671n ± 11%   4.714n ±  8%  -45.63% (p=0.000 n=9)
128Branches/16    5.731n ± 22%   5.388n ± 13%        ~ (p=0.474 n=9)
64Sizes/32       11.040n ± 17%   7.907n ± 13%  -28.38% (p=0.000 n=9)
64Sizes/64        15.77n ± 29%   11.72n ±  6%  -25.68% (p=0.000 n=9)
64Sizes/128       25.79n ± 30%   19.96n ±  6%  -22.61% (p=0.001 n=9)
64Sizes/256       40.22n ± 62%   36.51n ±  2%   -9.22% (p=0.002 n=9)
64Sizes/512       73.45n ± 67%   68.54n ±  4%        ~ (p=0.489 n=9)
64Sizes/1024      137.4n ± 62%   140.8n ±  2%        ~ (p=0.546 n=9)
64Sizes/2048      268.9n ± 12%   263.4n ±  2%        ~ (p=0.154 n=9)
64Sizes/4096      527.4n ±  9%   534.7n ±  2%        ~ (p=0.436 n=9)
64Sizes/8192      1.010µ ±  5%   1.058µ ±  2%   +4.75% (p=0.047 n=9)
128Sizes/32       7.759n ±  5%   7.267n ±  4%   -6.34% (p=0.004 n=9)
128Sizes/64       11.43n ±  6%   11.09n ±  5%        ~ (p=0.094 n=9)
128Sizes/128      19.61n ±  3%   19.46n ±  4%        ~ (p=0.605 n=9)
128Sizes/256      34.55n ±  6%   36.00n ±  7%        ~ (p=0.136 n=9)
128Sizes/512      64.97n ±  6%   69.94n ±  5%   +7.65% (p=0.006 n=9)
128Sizes/1024     128.9n ±  5%   140.9n ±  8%   +9.31% (p=0.002 n=9)
128Sizes/2048     254.1n ± 11%   271.7n ±  4%        ~ (p=0.113 n=9)
128Sizes/4096     498.7n ± 13%   547.3n ±  5%   +9.75% (p=0.040 n=9)
128Sizes/8192     1.002µ ± 10%   1.065µ ±  9%   +6.29% (p=0.029 n=9)
geomean           26.41n         22.88n        -13.37%

               │ rr_gen_c.txt  │             rr_new_c.txt              │
```

