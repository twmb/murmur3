murmur3
=======

Native Go implementation of Austin Appleby's third MurmurHash revision (aka
MurmurHash3).

Includes assembly for amd64 (go 1.5+), the benchmarks of which can be seen in
PR [#1](https://github.com/twmb/murmur3/pull/1).

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
