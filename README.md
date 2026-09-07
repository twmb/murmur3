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

Per 16 byte block that is 8.2 cycles for the assembly and 8.1 for the Go
loop on Go 1.27, and perf's port counters say what governs it. The block loop
issues 23 instructions per block: two loads, four multiplies, four rotates,
two xors, five LEAs, one ten byte immediate, one counter add and a fused
compare and branch. The four multiplies can only run on port 1 and the LEAs
only on ports 1 and 5, so port 1 is the contended resource, and the loop
carried h1 to h2 chain is five cycles because the compiler folds 5*c1 + c2
into a single constant, which takes an add off that chain. The assembly is
19 instructions with slack on every port and sits on its own eight cycle
chain, two three operand LEAs at three cycles each.

Three things in the source were chosen against measurement. The loop is an
index loop over data[i:i+16] with the bound hoisted by hand rather than a
reslicing loop: the compiler proves the window from the bound, folds the
index into the loads, and advances one counter instead of a pointer, a length
and a capacity, which is three fewer instructions and two fewer LEAs per
block. The two multipliers are loaded once from variables rather than written
as constants, because the compiler rematerializes a constant with a ten byte
MOVQ at every use inside a loop. And nothing else: unrolling, folding the
adds by hand so the folded constant could live in a register too, moving the
loop into its own function, and pointer arithmetic through unsafe were all
measured and lost, by lengthening the chain, by handing the allocator a
problem it solved with LEAs on the contended port, or in unsafe's case by
the checks unsafe.Slice inserts.

Byte identical loops can still differ by 15 percent, and perf showed why. Go
1.26 and 1.27 emit this exact loop with one register named differently, R13
against CX. R13 needs a REX prefix, so the loop is three bytes longer, the
micro op cache groups it differently for the allocator, and the port binding
heuristic puts one more micro op per block on port 1: 9.3 cycles per block on
Go 1.26 against 8.1 on 1.27. Go 1.25 cannot prove the window in bounds and
spills, 10.7. The reslicing loop this replaced measured 8.5 on both 1.26 and
1.27. This shape is kept because its floor is lower on every compiler from
1.26 on, and which register name a release picks is not something source
controls. GOAMD64=v3 changes nothing.

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
algorithm changes that. The index loop, which has no reslice guard, is what
it gains at 32 to 64 bytes.
