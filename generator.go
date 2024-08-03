package generator

import (
	"context"
)

// Yield is a function defined by the consumer and called by the producer of the generator.
// The first argument is the value that the producer passes to the consumer.
// The second argument is the error that may occur when generating the value.
// If the error is not nil, the consumer should decide whether to stop generation or continue.
// The return value is a flag that tells whether the generation should continue or stop.
// If the return value is false, then the generation stops.
type Yield[T any] func(T, error) bool

// Generator is simply a function that takes a Yield function from the consumer and returns an error.
type Generator[OUT any] func(Yield[OUT])

// Mapper is a function that transforms one value into another.
type Mapper[IN any, OUT any] func(IN, error) (OUT, error)

// Map takes a generator and a transformer function that will be applied to each value,
// generated by the generator. The transformer function changes the type of the value and value itself.
func Map[IN any, OUT any](generator Generator[IN], mapper Mapper[IN, OUT]) Generator[OUT] {
	return func(yield Yield[OUT]) {
		generator(func(v IN, err error) bool {
			return yield(mapper(v, err))
		})
	}
}

// MMap takes a generator and a transformer function into an array, which will be applied to each value,
// generated by the generator. Unlike Map, MMap can change not only types but also the number of values.
func MMap[IN any, OUT any](generator Generator[IN], mapper Mapper[IN, []OUT]) Generator[OUT] {
	return func(yield Yield[OUT]) {
		generator(func(v IN, err error) bool {
			outs, err := mapper(v, err)
			for _, out := range outs {
				if !yield(out, err) {
					return false
				}
			}
			return true
		})
	}
}

// Collect collects all values generated by the generator into a slice.
// In case of error transmission from the producer, the consumer will not be able to receive and handle it, so the generation immediately ends.
func (g Generator[OUT]) Collect() []OUT {
	var result []OUT
	g(func(v OUT, err error) bool {
		if err != nil {
			return false
		}
		result = append(result, v)
		return true
	})
	return result
}

// Seq transforms the generator into a rangefunc. This allows you to use the generator in a for loop in go >= 1.23.
// In case of error transmission from the producer, the consumer will not be able to receive and handle it, so the generation immediately ends.
func (g Generator[OUT]) Seq() func(yield func(OUT) bool) {
	return func(yield func(OUT) bool) {
		g(func(v OUT, err error) bool {
			if err != nil {
				return false
			}
			return yield(v)
		})
	}
}

// Chan transforms the generator into a channel.
// Since the generator can generate values indefinitely,
// a context is required to stop generation and close the channel.
// In case of error transmission from the producer, the consumer will not be able to receive and handle it, so the generation immediately ends.
// To stop generation, the consumer must close the context.
// If you exit the iteration and do not close the context, the channel will not be closed and the goroutine will hang.
func (g Generator[OUT]) Chan(ctx context.Context) <-chan OUT {
	var out = make(chan OUT)
	go func() {
		defer close(out)
		g(func(v OUT, err error) bool {
			if err != nil {
				return false
			}
			select {
			case <-ctx.Done():
				return false
			default:
				out <- v
			}
			return true
		})
	}()
	return out
}
