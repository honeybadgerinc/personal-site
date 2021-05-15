package main

import (
	"testing"
)

const (
	length = 100
)

type Inner struct {
	a int
	b string
	c []byte
}

type Outer struct {
	a int
	b string
	c []*Inner
}

func BenchmarkMapIterations(b *testing.B) {
	m := createMap()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for key, value := range m {
			_ = key
			_ = value
		}
	}
}

func BenchmarkSliceIterations(b *testing.B) {
	m := createSlice()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for idx, item := range m {
			_ = idx
			_ = item
		}
	}
}

func BenchmarkCreateMapConstruction(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = createMap()
	}
}

func BenchmarkCreateSliceConstruction(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = createSlice()
	}
}

func BenchmarkBasicMapConstruction(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = map[uint64]*Outer{}
	}
}

func BenchmarkBasicSliceConstruction(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = []*Outer{}
	}
}

func BenchmarkMakeMapConstruction(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = make(map[uint64]*Outer, length)
	}
}

func BenchmarkMakeSliceConstruction(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = make([]*Outer, length)
	}
}

func BenchmarkAppendToSliceFromMap(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		src := createMap()
		dst := make([]*Outer, len(src))
		b.StartTimer()
		for _, inst := range src {
			dst = append(dst, inst)
		}
	}
}

func BenchmarkAppendToSliceFromSlice(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		src := createSlice()
		dst := make([]*Outer, len(src))
		b.StartTimer()
		for _, inst := range src {
			dst = append(dst, inst)
		}
	}
}

func BenchmarkInsertIntoSliceFromMap(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		src := createMap()
		dst := make([]*Outer, len(src))
		b.StartTimer()
		for idx, inst := range src {
			dst[idx] = inst
		}
	}
}

func BenchmarkInsertIntoSliceFromSlice(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		src := createSlice()
		dst := make([]*Outer, len(src))
		b.StartTimer()
		for idx, inst := range src {
			dst[idx] = inst
		}
	}
}

func createSlice() []*Outer {
	s := make([]*Outer, length)
	for i := 0; i < length; i++ {
		s[i] = &Outer{
			a: 0,
			b: "",
			c: []*Inner{
				{
					a: 0,
					b: "",
					c: []byte{},
				},
			},
		}
	}
	return s
}

func createMap() map[int]*Outer {
	m := make(map[int]*Outer, length)
	for i := 0; i < length; i++ {
		m[i] = &Outer{
			a: 0,
			b: "",
			c: []*Inner{
				{
					a: 0,
					b: "",
					c: []byte{},
				},
			},
		}
	}
	return m
}
