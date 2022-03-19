package data

import (
	"go4ml.xyz/lazy"
	"testing"
)

var colors_N = func(N int) []Color {
	r := make([]Color, len(colors)*N)
	for i := 0; i < N; i++ {
		copy(r[len(colors)*i:len(colors)*(i+1)], colors)
	}
	return r
}(100)

func Benchmark_TableFromChannel(b *testing.B) {
	for n := 0; n < b.N; n++ {
		c := make(chan Color)
		go func() {
			for _, x := range colors_N {
				c <- x
			}
			close(c)
		}()
		var x Table
		lazy.Chan(c).Map1(StructToRow).MustDrain(x.Sink())
	}
}

func Benchmark_TableFromBigList(b *testing.B) {
	var x Table
	for i := 0; i < b.N; i++ {
		lazy.List(colors_N).Map1(StructToRow).MustDrain(x.Sink(len(colors_N)))
	}
}

