murmur3
=======

Native Go implementation of Austin Appleby's third MurmurHash revision (aka
MurmurHash3).

Includes assembly for amd64 for go1.5+ for 128 bit hashes, seeding function,
and string functions to avoid string to slice conversions.

Hand rolled 32 bit assembly was removed during 1.11 due to Go's compiler
catching up and generating equal or better assembly.

The reference algorithm has been slightly hacked as to support the streaming mode
required by Go's standard [Hash interface](http://golang.org/pkg/hash/#Hash).

Testing
=======

[![Build Status](https://travis-ci.org/twmb/murmur3.svg?branch=master)](https://travis-ci.org/twmb/murmur3)

Testing includes comparing random inputs against the [canonical
implementation](https://github.com/aappleby/smhasher/blob/master/src/MurmurHash3.cpp),
and testing length 0 through 17 inputs to force all branches.

Documentation
=============

[![GoDoc](https://godoc.org/github.com/twmb/murmur3?status.svg)](https://godoc.org/github.com/twmb/murmur3)

Full documentation can be found on `godoc`.

Benchmarks
==========

The following benchmarks show deltas for the 128 bit algorithms only; the 32
bit algorithms have the same implementation.

In comparison to [spaolacci/murmur3](https://github.com/spaolacci/murmur3) on
Go at commit [447965d4e0](https://github.com/golang/go/commit/447965d4e0)
(i.e., post 1.11):

```
benchmark                     old ns/op     new ns/op     delta
Benchmark128Branches/0-4      22.2          6.28          -71.71%
Benchmark128Branches/1-4      23.6          8.46          -64.15%
Benchmark128Branches/2-4      24.3          8.68          -64.28%
Benchmark128Branches/3-4      24.7          9.07          -63.28%
Benchmark128Branches/4-4      25.2          8.16          -67.62%
Benchmark128Branches/5-4      25.9          8.89          -65.68%
Benchmark128Branches/6-4      26.8          9.32          -65.22%
Benchmark128Branches/7-4      27.4          9.82          -64.16%
Benchmark128Branches/8-4      28.1          7.68          -72.67%
Benchmark128Branches/9-4      29.6          9.04          -69.46%
Benchmark128Branches/10-4     30.2          9.14          -69.74%
Benchmark128Branches/11-4     30.8          9.53          -69.06%
Benchmark128Branches/12-4     31.5          8.65          -72.54%
Benchmark128Branches/13-4     31.5          9.26          -70.60%
Benchmark128Branches/14-4     32.5          9.69          -70.18%
Benchmark128Branches/15-4     33.4          10.1          -69.76%
Benchmark128Branches/16-4     24.9          10.0          -59.84%
Benchmark64Sizes/32-4         27.8          13.6          -51.08%
Benchmark64Sizes/64-4         35.2          18.8          -46.59%
Benchmark64Sizes/128-4        49.6          30.5          -38.51%
Benchmark64Sizes/256-4        77.9          54.5          -30.04%
Benchmark64Sizes/512-4        136           105           -22.79%
Benchmark64Sizes/1024-4       251           209           -16.73%
Benchmark64Sizes/2048-4       492           419           -14.84%
Benchmark64Sizes/4096-4       952           832           -12.61%
Benchmark64Sizes/8192-4       1879          1658          -11.76%
Benchmark128Sizes/32-4        28.5          13.6          -52.28%
Benchmark128Sizes/64-4        35.7          18.7          -47.62%
Benchmark128Sizes/128-4       49.8          30.3          -39.16%
Benchmark128Sizes/256-4       78.0          54.2          -30.51%
Benchmark128Sizes/512-4       135           105           -22.22%
Benchmark128Sizes/1024-4      250           209           -16.40%
Benchmark128Sizes/2048-4      489           419           -14.31%
Benchmark128Sizes/4096-4      959           831           -13.35%
Benchmark128Sizes/8192-4      1885          1659          -11.99%
BenchmarkNoescape128-4        3226          1824          -43.46%
```

The speedup for large inputs levels out around ~1.12x. Additionally,
this code avoids allocating stack slices unnecessarily for the 128
algorithm, unlike `spaolacci/murmur3`.
