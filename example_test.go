package generator

import (
	"context"
	"fmt"
)

func gen1M() Generator[int] {
	return func(yield Yield[int]) {
		for i := 1; i <= 1_000_000; i++ {
			if !yield(i, nil) {
				return
			}
		}
	}
}

func genSeq1M() func(yield func(int) bool) {
	return func(yield func(int) bool) {
		for i := 1; i <= 1_000_000; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

func gen1MChan() <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; i < 1_000_000; i++ {
			out <- i
		}
	}()
	return out
}

func multiply[T ~int | ~int64 | ~uint64](mul T) Mapper[T, T] {
	return func(t T, err error) (T, error) {
		return t * mul, err
	}
}

func toFloat[T ~int | ~int64 | ~uint64]() Mapper[T, float64] {
	return func(t T, err error) (float64, error) {
		return float64(t) + 0.1, err
	}
}

func findMinus() Yield[float64] {
	return func(v float64, err error) bool {
		if v < 0 {
			fmt.Printf("new one digit: %f", v)
		}
		return true
	}
}

func exampleSyncGenerator() {
	ints := gen1M()
	mul := Map(ints, multiply(2))
	floats := Map(mul, toFloat[int]())
	floats(findMinus())
}

func exampleChan() {
	ints := gen1M()
	in := ints.Chan(context.Background())
	for v := range in {
		if v < 0 {
			fmt.Printf("new one digit: %d", v)
		}
	}
}
