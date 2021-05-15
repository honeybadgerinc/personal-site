---
title: Benchmarking Go's data types
date: "2021-05-14"
description: Comparing the performance of Go's data structures
tags: [go, programming]
---

I've been thinking a lot lately about the performance of data structures like maps, arrays, slices, etc., in Go. Obviously,
a map will have O(1) insertions and lookups; a list has O(n) lookups; [other CS 101 stuff]. I'm more interested in knowing
how those data structures perform on real hardware.

To that end, I wrote benchmarks for operations like creating a slice or map; iterating over a map's values;
creation with `make` rather than `var x []Foo`; whether it's faster to put map values into a slice (by traversing the map), 
or faster to traverse a slice and append its values that way, etc.

I strongly encourage readers to view the source: [bench_test.go](/static/bench_test.go). I will give only a high-level
overview of the results below; the reader should look at the code and see what the benchmark functions actually are.

### tldr: nothing here is surprising

```bash
$ go test -bench . content/static/bench_test.go
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i5-5200U CPU @ 2.20GHz // I know this is sad

BenchmarkMapIterations-4                  739734     1524  ns/op    0 B/op       0 allocs/op
BenchmarkSliceIterations-4              20021416     65.10 ns/op    0 B/op       0 allocs/op
BenchmarkMapConstruction-4                 56673     20438 ns/op    13239 B/op   302 allocs/op
BenchmarkSliceConstruction-4               81038     17725 ns/op    11296 B/op   301 allocs/op
BenchmarkBasicMapConstruction-4        155245270     7.567 ns/op    0 B/op       0 allocs/op
BenchmarkBasicSliceConstruction-4     1000000000     0.376 ns/op    0 B/op       0 allocs/op
BenchmarkMakeMapConstruction-4            977562     1113  ns/op    2712 B/op    2 allocs/op
BenchmarkMakeSliceConstruction-4         3097621     371.2 ns/op    896 B/op     1 allocs/op
BenchmarkAppendToSliceFromMap-4           308476     3988  ns/op    1792 B/op    1 allocs/op
BenchmarkAppendToSliceFromSlice-4         946756     1311  ns/op    1792 B/op    1 allocs/op
BenchmarkInsertIntoSliceFromMap-4         440878     2816  ns/op    0 B/op       0 allocs/op
BenchmarkInsertIntoSliceFromSlice-4      4340052     275.1 ns/op    0 B/op       0 allocs/op
```

### Summary

* Iterating
  * Traversing slices is much, much faster than traversing maps.
* Constructing
  * Constructing a slice via `f := []*Foo` is faster than invoking `make([]*Foo, len(length))`, and both are faster than constructing 
  a similar map (either way).
  * Creating a map with `make` has two allocations. Creating a slice in a similar way has only one.
* Appending
  * Appending to a slice from another slice is faster than appending to a slice from a map.
* Inserting
  * This one was interesting. Notice that it took the benchmark tool about four million invocations to get an answer about slices, where
  maps only needed about 400,000 invocations. I have no idea why that is. But again, inserting data from a slice into another slice
  is much faster than the same operation from a map into a slice.

Please feel free to comment or critique the benchmarks. Or, use them in your development environment with your business objects
(you'll obviously have to do some editing of the source) to see how they perform under stress. More data is always more helpful
than less data, and you should never, under any circumstances, have to guess how performant your business objects will be. 

And always remember: "...kids, the only difference between science and screwing around is writing it down."