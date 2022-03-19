package data

import (
	"go4ml.xyz/lazy"
	"testing"
)

func bench_transform_1(interface{}) interface{} {
	type S struct{ Value int }
	w, _ := NewWrapper(S{})
	return func(x interface{}) interface{} {
		y := x.(Color).Index
		for j := 0; j < 60000; j++ {
			y = y * x.(Color).Index
		}
		return w.WrapOrFail(S{y})
	}
}

func Benchmark_TableFromBigListLinear(b *testing.B) {
	var x Table
	for i := 0; i < b.N; i++ {
		lazy.List(colors_N).Map1(bench_transform_1).MustDrain(x.Sink(len(colors_N)))
	}
}

func Benchmark_TableFromBigGenLinear(b *testing.B) {
	var x Table
	for i := 0; i < b.N; i++ {
		lazy.Generator(func(j int) (interface{}) {
			if j >= len(colors_N) {
				return lazy.EoS
			}
			return colors_N[j]
		}).Map1(bench_transform_1).MustDrain(x.Sink(len(colors_N)))
	}
}

func Benchmark_TableFromBigSeqLinear(b *testing.B) {
	var x Table
	for i := 0; i < b.N; i++ {
		lazy.Sequence(func(j int) (interface{}) {
			if j >= len(colors_N) {
				return lazy.EoS
			}
			return colors_N[j]
		}).Map1(bench_transform_1).MustDrain(x.Sink(len(colors_N)))
	}
}

func Benchmark_TableFromBigListCcr(b *testing.B) {
	var x Table
	for i := 0; i < b.N; i++ {
		lazy.List(colors_N).Map1(bench_transform_1).MustDrain(x.Sink(len(colors_N)), 8)
	}
}

func Benchmark_TableFromBigGenCcr(b *testing.B) {
	var x Table
	for i := 0; i < b.N; i++ {
		lazy.Generator(func(j int) interface{} {
			if j >= len(colors_N) {
				return lazy.EoS
			}
			return colors_N[j]
		}).Map1(bench_transform_1).MustDrain(x.Sink(len(colors_N)), 8)
	}
}

func Benchmark_TableFromBigSeqCcr(b *testing.B) {
	var x Table
	for i := 0; i < b.N; i++ {
		lazy.Sequence(func(j int) (interface{}) {
			if j >= len(colors_N) {
				return lazy.EoS
			}
			return colors_N[j]
		}).Map1(bench_transform_1).MustDrain(x.Sink(len(colors_N)), 8)
	}
}
