package generator

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

var generatorClosed = false

func genTest(yield Yield[string]) {
	generatorClosed = false
	defer func() {
		generatorClosed = true
	}()

	if !yield("value1", nil) || !yield("value2", nil) || !yield("value3", nil) {
		return
	}
}
func genTest2E(yield Yield[string]) {
	generatorClosed = false
	defer func() {
		generatorClosed = true
	}()

	err := errors.New("some error")

	if !yield("value1", nil) || !yield("value2", err) || !yield("value3", nil) {
		return
	}
}

func chanToSlice[OUT any](g Generator[OUT], limit int) []OUT {
	var result []OUT
	ind := 0
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if 0 == limit {
		cancel()
	}
	for v := range g.Chan(ctx) {
		result = append(result, v)
		ind += 1
		if ind == limit {
			cancel()
		}
	}
	return result
}

func seqToSlice[OUT any](g Generator[OUT], limit int) []OUT {
	var result []OUT
	ind := 0
	g.Seq()(func(v OUT) bool {
		result = append(result, v)
		ind += 1
		if ind == limit {
			return false
		}
		return true
	})
	// requires go >=1.23
	//for v := range g.Seq() {
	//	result = append(result, v)
	//	ind += 1
	//	if ind == limit {
	//		break
	//	}
	//}
	return result
}

func TestGenerator_Chan(t *testing.T) {
	g := genTest
	type testCase[OUT any] struct {
		name  string
		g     Generator[OUT]
		limit int
		want  []OUT
	}
	tests := []testCase[string]{
		{
			name:  "chan all values",
			g:     g,
			limit: 3,
			want:  []string{"value1", "value2", "value3"},
		},
		{
			name:  "chan two values",
			g:     g,
			limit: 2,
			want:  []string{"value1", "value2"},
		},
		{
			name:  "chan no one values",
			g:     g,
			limit: 0,
			want:  nil,
		},
		{
			name:  "generator with error",
			g:     genTest2E,
			limit: 3,
			want:  []string{"value1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chanToSlice(tt.g, tt.limit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Chan() = %v, want %v", got, tt.want)
			}
			if !generatorClosed {
				t.Errorf("Generator should be closed")
			}
		})
	}
}

func TestGenerator_Seq(t *testing.T) {
	g := genTest
	type testCase[OUT any] struct {
		name  string
		g     Generator[OUT]
		limit int
		want  []OUT
	}
	tests := []testCase[string]{
		{
			name:  "seq all values",
			g:     g,
			limit: 3,
			want:  []string{"value1", "value2", "value3"},
		},
		{
			name:  "seq two values",
			g:     g,
			limit: 2,
			want:  []string{"value1", "value2"},
		},
		{
			name:  "seq one values",
			g:     g,
			limit: 1,
			want:  []string{"value1"},
		},
		{
			name:  "generator with error",
			g:     genTest2E,
			limit: 3,
			want:  []string{"value1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := seqToSlice(tt.g, tt.limit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Seq() = %v, want %v", got, tt.want)
			}
			if !generatorClosed {
				t.Errorf("Generator should be closed")
			}
		})
	}
}

func TestGenerator_Collect(t *testing.T) {
	type testCase[OUT any] struct {
		name string
		g    Generator[OUT]
		want []OUT
	}
	tests := []testCase[string]{
		{
			name: "collect all values",
			g:    genTest,
			want: []string{"value1", "value2", "value3"},
		},
		{
			name: "generator with error",
			g:    genTest2E,
			want: []string{"value1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.Collect(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Collect() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMap(t *testing.T) {
	g := func(yield Yield[int]) {
		generatorClosed = false
		defer func() {
			generatorClosed = true
		}()

		if !yield(1, nil) || !yield(2, nil) || !yield(3, nil) {
			return
		}
	}
	type args[IN any, OUT any] struct {
		generator Generator[IN]
		mapper    Mapper[IN, OUT]
	}
	type testCase[IN any, OUT any] struct {
		name string
		args args[IN, OUT]
		want []OUT
	}
	tests := []testCase[int, string]{
		{
			name: "map int to string",
			args: args[int, string]{
				generator: g,
				mapper: func(t int, err error) (string, error) {
					return fmt.Sprintf("%d", t), err
				},
			},
			want: []string{"1", "2", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Map(tt.args.generator, tt.args.mapper).Collect(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMMap(t *testing.T) {
	type args[IN any, OUT any] struct {
		generator Generator[IN]
		mapper    Mapper[IN, []OUT]
	}
	type testCase[IN any, OUT any] struct {
		name  string
		args  args[IN, OUT]
		limit int
		want  []OUT
	}
	tests := []testCase[int, string]{
		{
			name: "map all int to string",
			args: args[int, string]{
				generator: func(yield Yield[int]) {
					generatorClosed = false
					defer func() {
						generatorClosed = true
					}()

					if !yield(1, nil) || !yield(2, nil) || !yield(3, nil) {
						return
					}
				},
				mapper: func(t int, err error) ([]string, error) {
					result := make([]string, t)
					for n := 0; n < t; n++ {
						result[n] = fmt.Sprintf("%d", t)
					}
					return result, err
				},
			},
			limit: 6,
			want:  []string{"1", "2", "2", "3", "3", "3"},
		},
		{
			name: "map only 4 int to string",
			args: args[int, string]{
				generator: func(yield Yield[int]) {
					generatorClosed = false
					defer func() {
						generatorClosed = true
					}()

					if !yield(1, nil) || !yield(2, nil) || !yield(3, nil) {
						return
					}
				},
				mapper: func(t int, err error) ([]string, error) {
					result := make([]string, t)
					for n := 0; n < t; n++ {
						result[n] = fmt.Sprintf("%d", t)
					}
					return result, err
				},
			},
			limit: 4,
			want:  []string{"1", "2", "2", "3"},
		},
		{
			name: "map no one int to string",
			args: args[int, string]{
				generator: func(yield Yield[int]) {
					generatorClosed = false
					defer func() {
						generatorClosed = true
					}()

					if !yield(1, nil) || !yield(2, nil) || !yield(3, nil) {
						return
					}
				},
				mapper: func(t int, err error) ([]string, error) {
					result := make([]string, t)
					for n := 0; n < t; n++ {
						result[n] = fmt.Sprintf("%d", t)
					}
					return result, err
				},
			},
			limit: 0,
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chanToSlice(MMap(tt.args.generator, tt.args.mapper), tt.limit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
