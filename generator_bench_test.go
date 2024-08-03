package generator

import (
	"fmt"
	"testing"
)

func BenchmarkEmptyGenerator(b *testing.B) {
	generator := func(yield Yield[int]) error { return nil }
	yield := func(v int, err error) bool { return true }
	for i := 0; i < b.N; i++ {
		err := generator(yield)
		if err != nil {
			return
		}
	}
}

func BenchmarkEmptyChannels(b *testing.B) {
	g := func() <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for i := 0; i < 0; i++ {
				out <- i
			}
		}()
		return out
	}()
	yield := func(v int) bool { return true }
	for i := 0; i < b.N; i++ {
		for v := range g {
			if !yield(v) {
				break
			}
		}
	}
}

func Benchmark1MIterates_Generator(b *testing.B) {
	g := gen1M()
	yield := func(v int, err error) bool { return true }
	for i := 0; i < b.N; i++ {
		g(yield)
	}
}

func Benchmark1MIterates_Channel(b *testing.B) {
	g := gen1MChan()
	yield := func(v int) bool { return true }
	for i := 0; i < b.N; i++ {
		for v := range g {
			if !yield(v) {
				break
			}
		}
	}
}

func Benchmark1MIterates_Seq(b *testing.B) {
	g := genSeq1M()
	yield := func(v int) bool { return true }
	for i := 0; i < b.N; i++ {
		g(yield)
	}
}

func BenchmarkPipe_Generator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		exampleSyncGenerator()
	}
}

func BenchmarkPipe_Channel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		exampleChannel()
	}
}

func BenchmarkPipe_Plain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		examplePlain()
	}
}

func BenchmarkPipe_PlainCompact(b *testing.B) {
	for i := 0; i < b.N; i++ {
		examplePlainCompact()
	}
}

func BenchmarkPipe_Plain_WithoutPreAllocation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		examplePlainWithoutPreAllocation()
	}
}

func examplePlain() {
	// generate
	var ints [1_000_000]int
	for i := 0; i < 1_000_000; i++ {
		ints[i] = i
	}

	// multiply
	var mul [1_000_000]int
	for i := 0; i < 1_000_000; i++ {
		mul[i] = ints[i] * 2
	}

	// to float
	var flts [1_000_000]float64
	for i := 0; i < 1_000_000; i++ {
		flts[i] = float64(mul[i]) + 0.1
	}

	// format
	for i := 0; i < 1_000_000; i++ {
		if flts[i] < 0 {
			fmt.Printf("new one digit: %f", flts[i])
		}
	}
}

func exampleChannel() {
	gen := gen1MChan()

	mul := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for i := range in {
				out <- i * 2
			}
		}()
		return out
	}

	flt := func(in <-chan int) <-chan float64 {
		ch := make(chan float64)
		go func() {
			defer close(ch)
			for i := range in {
				ch <- float64(i) + 0.1
			}
		}()
		return ch
	}

	check := func(in <-chan float64) {
		for i := range in {
			if i < 0 {
				fmt.Printf("new one digit: %f", i)
			}
		}
	}

	check(flt(mul(gen)))
}

func examplePlainCompact() {
	for i := 0; i < 1_000_000; i++ {
		flt := float64(i * 2)
		if flt < 0 {
			fmt.Printf("new one digit: %f", flt)
		}
	}
}

func examplePlainWithoutPreAllocation() {
	// generate
	ints := make([]int, 0)
	for i := 0; i < 1_000_000; i++ {
		ints = append(ints, i)
	}

	// multiply
	mul := make([]int, 0)
	for i := 0; i < 1_000_000; i++ {
		mul = append(mul, ints[i]*2)
	}

	// to float
	flts := make([]float64, 0)
	for i := 0; i < 1_000_000; i++ {
		flts = append(flts, float64(mul[i])+0.1)
	}

	// format
	for i := 0; i < 1_000_000; i++ {
		if flts[i] < 0 {
			fmt.Printf("new one digit: %f", flts[i])
		}
	}
}
