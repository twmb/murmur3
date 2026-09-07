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
under 64 bytes, for keys of varying length, and for every 32 bit sum, and 3
to 9 percent slower on fixed inputs of 256 bytes and more. The 32 bit
assembly was removed for the same reason back in Go 1.11. See the benchmarks
below.

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
tail; the `RandomLengths` rows draw a fresh xorshift length in [0, 64) on
every call, which is what a real key stream looks like to the branch
predictor. Each row includes a few cycles of benchmark harness.

```
benchmark                asm  old Go  new Go   new/asm new/old
128Sizes/8192           4213    4479    4387     +4.1%   -2.1%
128Sizes/1024            534     574     569     +6.6%   -0.9%
128Sizes/256             140     156     152     +8.6%   -2.9%
128Sizes/64               48      52      49     +1.4%   -7.1%
128Sizes/32               35      36      34     -2.6%   -4.2%
128Branches/16            26      23      25     -5.1%   +6.6%
128Branches/13            25      34      21    -14.4%  -37.0%
128Branches/7             26      26      21    -18.2%  -19.7%
128Branches/3             25      22      21    -16.5%   -3.0%
64Sizes/8192            4215    4530    4332     +2.8%   -4.4%
64Sizes/32                35      38      34     -3.1%  -10.2%
32Sizes/8192           10074   10045    9897     -1.8%   -1.5%
32Sizes/64                88      88      78    -11.5%  -11.5%
32Sizes/32                50      50      42    -14.6%  -14.5%
32Branches/3              15      15      16     +5.6%   +5.7%
RandomLengths128         110      96      89    -19.0%   -6.8%
RandomLengths32           95      96      87     -8.8%   -9.3%
```

Per 16 byte block that is 8.2 cycles for the assembly and 8.5 for the Go
loop, and perf's port counters say where the difference is. The Go loop
issues 29 micro ops per block to the assembly's 19, so on a four wide core its
floor is 7.3 cycles and ports 0, 1 and 6 each run 75 to 80 percent busy; the
measured 8.5 is scheduling slack on top of that. The assembly has slack on
every port and sits on its dependency chain instead: the four 64 bit
multiplies and the rotates are off that chain, and what is left is h1 to h2
through two three operand LEAs at three cycles each. The one loop change that
helped was the strictly greater bound, which removes the zero length pointer
guard the compiler emits after every reslice, four micro ops per block,
without touching the arithmetic. Everything else was measured and lost:
unrolling two or four blocks per iteration, hoisting the constants into
registers, folding the two adds by hand, and moving the loop into its own
function. Cutting micro ops only wins if the h1 to h2 chain stays at the five
cycles the compiler's own folding of 5*c1 + c2 gives it, and which registers
the allocator hands out decides whether a three address add is an ADD or an
LEA. Byte identical loops landed anywhere from 8.5 to 11 cycles per block on
that alone. That also makes the result a property of the compiler version:
Go 1.26 and 1.27 both produce the 8.5 cycle loop from this source, while Go
1.25 emits four more instructions for it and lands at 9.5. The best shape
found on Go 1.27, an index loop over data[i:i+16] with the two multipliers
loaded once into locals, measures 8.1 cycles per block there, but 9.2 on Go
1.26 and 10.9 on Go 1.25, so it is not used: this loop is the one that is
stable across the supported compilers. GOAMD64=v3 changes nothing.

The tail is where the two differ most, in both directions. On a fixed length
every branch in a tail is perfectly predicted, and the Go code wins by
reading the tail in two or three word sized loads where the assembly walks a
tree of byte loads. On varying lengths a switch on the tail length is a jump
table, an indirect branch that mispredicts nearly every time the length
changes, and each miss is about 20 cycles: on a mixed stream the first
version of this code took 0.94 mispredicts per hash to the assembly's 0.30
and was 34 percent slower. So once an input has had a full block, the tail is
read out of the input's last 16 bytes, which are always in bounds, with two
overlapping loads and shifts that the length selects without a branch, and
this happens before the block loop so that only two words stay live across
it. Only inputs shorter than one block take the switch. That is what the
RandomLengths rows measure: 19 percent under the assembly for 128 bit sums
and 9 percent for 32 bit. Sum64 and its variants call the shared
implementation directly rather than through Sum128, which kept them under
the inliner's budget and saves a call per hash.

The 32 bit sum is bound by its own four cycle per block dependency chain and
sits at 4.8 with every port under 80 percent busy; nothing above the
algorithm changes that. The strict bound is what it gains at 32 to 64 bytes.
