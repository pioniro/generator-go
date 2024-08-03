/*
Package generator provides the ability to create generators and transform them.
It provides the Generator and Yield types, as well as the Map, MMap, Collect, Seq, and Chan functions.
A simple usage example:

	package main

	func generateNumbers(yield Yield[int]) {
	  for i := 0; i < 10; i++ {
	    if !yield(i, nil) {
	      return
	    }
	  }
	}

	func main() {
	  generateNumbers(func(v int, err error) bool {
	    fmt.Println(v)
	    // enough
	    if v == 5 {
	      return false
	    }
	    return true
	  })
	}

For type or value transformations, use Map:

	package main

	func generateNumbers(yield Yield[int]) {
	  for i := 0; i < 10; i++ {
	    if !yield(i, nil) {
	      return
	    }
	  }
	}

	func intToString(v int, err error) (string, error) {
	  return fmt.Sprint(v), err
	}

	func main() {
	  for v := range generateNumbers().Map(intToString).Collect() {
	    fmt.Println(v)
	  }
	}

For changing types, values, and the number of values, use MMap:

	package main

	func generateNumbers(yield Yield[int]) {
	  for i := 0; i < 10; i++ {
	    if !yield(i, nil) {
	      return
	    }
	  }
	}


	func intToStrings(v int, err error) ([]string, error) {
	  result := make([]string, i)
	  for n := 0; n < v; n++ {
	    result[n] = fmt.Sprint(v), err
	  }
	}

	func main() {
	  for v := range generateNumbers().Map(intToString).Chan() {
	    fmt.Println(v)
	  }
	}

To use the generator in a for loop, you can convert the values to a channel ( .Chan() ) or collect them into an array ( .Collect() ).
When using go >= 1.23, you can use Seq to convert the generator into a rangefunc function.
*/
package generator
