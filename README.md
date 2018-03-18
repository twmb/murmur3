murmur3
=======

Native Go implementation of Austin Appleby's third MurmurHash revision (aka
MurmurHash3).

Includes assembly for amd64 for go1.5+, and also includes string summing
functions to avoid string to slice conversions.

Reference algorithm has been slightly hacked as to support the streaming mode
required by Go's standard [Hash interface](http://golang.org/pkg/hash/#Hash).

Testing
=======

[![Build Status](https://travis-ci.org/twmb/murmur3.svg?branch=ci)](https://travis-ci.org/twmb/murmur3)

Testing includes comparing random inputs against the [canonical
implementation](https://github.com/aappleby/smhasher/blob/master/src/MurmurHash3.cpp),
and testing length 0 through 17 inputs to force all branches.

Documentation
=============

[![GoDoc](https://godoc.org/github.com/twmb/murmur3?status.svg)](https://godoc.org/github.com/twmb/murmur3)

Full documentation can be found on `godoc`.

Benchmarks
==========

In comparison to [spaolacci/murmur3](https://github.com/spaolacci/murmur3) on
Go at commit
[718d6c5880](https://github.com/golang/go/commit/718d6c5880fe3507b1d224789b29bc2410fc9da5)
(i.e., post 1.10):

```
benchmark                          old ns/op     new ns/op     delta
Benchmark32Branches/0-4            9.25          4.22          -54.38%
Benchmark32Branches/1-4            10.3          5.07          -50.78%
Benchmark32Branches/2-4            10.8          5.31          -50.83%
Benchmark32Branches/3-4            11.5          5.63          -51.04%
Benchmark32Branches/4-4            10.5          5.64          -46.29%
BenchmarkPartial32Branches/0-4     88.3          90.4          +2.38%
BenchmarkPartial32Branches/1-4     91.1          91.7          +0.66%
BenchmarkPartial32Branches/2-4     91.9          92.9          +1.09%
BenchmarkPartial32Branches/3-4     92.6          93.6          +1.08%
BenchmarkPartial32Branches/4-4     89.6          91.3          +1.90%
Benchmark128Branches/0-4           22.4          6.45          -71.21%
Benchmark128Branches/1-4           23.9          8.48          -64.52%
Benchmark128Branches/2-4           24.3          8.66          -64.36%
Benchmark128Branches/3-4           25.2          9.03          -64.17%
Benchmark128Branches/4-4           25.6          8.42          -67.11%
Benchmark128Branches/5-4           26.3          8.85          -66.35%
Benchmark128Branches/6-4           27.8          9.26          -66.69%
Benchmark128Branches/7-4           29.2          9.74          -66.64%
Benchmark128Branches/8-4           31.8          7.76          -75.60%
Benchmark128Branches/9-4           30.8          8.99          -70.81%
Benchmark128Branches/10-4          31.5          9.22          -70.73%
Benchmark128Branches/11-4          32.4          9.61          -70.34%
Benchmark128Branches/12-4          32.5          8.72          -73.17%
Benchmark128Branches/13-4          33.0          9.84          -70.18%
Benchmark128Branches/14-4          33.6          9.89          -70.57%
Benchmark128Branches/15-4          34.9          10.1          -71.06%
Benchmark128Branches/16-4          26.1          9.89          -62.11%
Benchmark32Sizes/32-4              24.1          16.2          -32.78%
Benchmark32Sizes/64-4              39.5          31.2          -21.01%
Benchmark32Sizes/128-4             73.1          55.4          -24.21%
Benchmark32Sizes/256-4             142           116           -18.31%
Benchmark32Sizes/512-4             287           251           -12.54%
Benchmark32Sizes/1024-4            565           508           -10.09%
Benchmark32Sizes/2048-4            1119          1026          -8.31%
Benchmark32Sizes/4096-4            2231          2054          -7.93%
Benchmark32Sizes/8192-4            4472          4109          -8.12%
Benchmark64Sizes/32-4              29.0          16.3          -43.79%
Benchmark64Sizes/64-4              35.9          21.6          -39.83%
Benchmark64Sizes/128-4             50.2          33.5          -33.27%
Benchmark64Sizes/256-4             81.8          57.8          -29.34%
Benchmark64Sizes/512-4             146           114           -21.92%
Benchmark64Sizes/1024-4            275           216           -21.45%
Benchmark64Sizes/2048-4            540           430           -20.37%
Benchmark64Sizes/4096-4            1067          851           -20.24%
Benchmark64Sizes/8192-4            2091          1698          -18.79%
Benchmark128Sizes/32-4             30.4          13.8          -54.61%
Benchmark128Sizes/64-4             37.1          18.8          -49.33%
Benchmark128Sizes/128-4            51.9          30.7          -40.85%
Benchmark128Sizes/256-4            82.5          55.3          -32.97%
Benchmark128Sizes/512-4            146           107           -26.71%
Benchmark128Sizes/1024-4           276           213           -22.83%
Benchmark128Sizes/2048-4           544           426           -21.69%
Benchmark128Sizes/4096-4           1056          847           -19.79%
Benchmark128Sizes/8192-4           2092          1695          -18.98%
```
