# fson - High-performance, Zero-Allocation JSON Encoder for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/LucasRouckhout/fson)](https://goreportcard.com/report/github.com/LucasRouckhout/fson)
[![GoDoc](https://godoc.org/github.com/LucasRouckhout/fson?status.svg)](https://godoc.org/github.com/LucasRouckhout/fson)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

`fson` is a high-performance JSON encoder for Go that focuses on simplicity and full control over heap allocations. 

- **Fluent API**: Simple chainable methods for building JSON structures
- **Complete Control**: Full control over the produced JSON and heap allocations.
- **Simple Implementation**: The entire core library is contained in a single file of ~1000 lines (mostly documentation)
- **Easy to Vendor**: The small codebase makes it easy to vendor `fson` into an existing codebase
- **No Reflection**: `fson` avoids reflection completely
- **Zero Allocations**: `fson` by itself will not allocate any memory on the heap, unless you force it to.

The finer details (especially UTF8 handling) of this library are heavily inspired by the json encoder of Uber's Zap logging library
[zapcore](https://github.com/uber-go/zap/tree/master/zapcore).

## Quick Example

```go
package main

import (
	"fmt"
	"github.com/LucasRouckhout/fson"
	"github.com/LucasRouckhout/fson/fsonutil"
)

var buffPool = fsonutil.NewPool()

func main() {
	buf := buffPool.Get()
	defer buffPool.Put(buf)

	b := fson.NewObject(buf.Bytes()).
		String("hello", "world").
		Bool("bool", true).
		Object("obj").
		String("foo", "bar").
		Bool("bool", false).
		Int("int", 8).
		EndObject().
		Build()

	fmt.Println(string(b)) // -> {"foo":"bar","bool":true,"obj":{"foo":"bar","bool":false,"int":8}}
}
```

# Usage

Add `fson` as a dependency

```
go get -u github.com/LucasRouckhout/fson
```

## Fluent API

`fson` exposes a "two-tiered" fluent API that enables method chaining to build up a JSON object. For example
this code snippet:

```go
fson.NewObject(buf).Key("foo").StringValue("bar").Build()
```

Would produce:

```json
{"foo": "bar"}
```

Or a more complex example

```go
fson.NewObject(buf).Key("fooArr").StartArray().StringValue("bar").StringValue("foo").EndArray().Build()
```

Would produce:

```json
{"fooArr":["bar","foo"]}
```

This explicit key-value approach can become cumbersome and verbose. So for most operations there exists a shorthand
"higher-level" API. These two examples can respectively be rewritten in a shorter manner like this.

```go
fson.NewObject(buf).String("foo", "bar").Build()
fson.NewObject(buf).Strings("fooArr", []string{"bar", "foo"}).Build()
```

For most use-cases the higher-level API will be enough. But there are examples, like multi-typed arrays, where you will
need to fall back to the lower level API to produce the desired output.

## A note on performance

The raison d'être for `fson` is to allow developers full control over both the produced JSON and heap allocations as 
much as possible. That's to say that while `fson` itself is very performant, using it incorrectly can cause you to lose 
all the performance gains it potentially offers. It's a classic case of: "With great power comes great responsibility".

> Incorrect use of `fson` will potentially negate all performance gains it has to offer.

Here are some tips and tricks to use `fson` in an efficient manner.

### Use the provided `fsonutil` buffer pool

Avoid using make inside a function that is called multiple times within the lifetime of your program.

```go
func IfYouCallThisFunctionALotThisIsBad() {
	// ...
	buff := make([]byte, 0, 1024)
	fson.NewObject(buff)
	// ...
}
```

This will allocate a new buffer on the heap everytime the function is called. Instead, you should be using a
buffer pool for this use-case. `fsonutil` provides a specialized buffer pool for use with `fson` that avoids
pinning large chunks of memory. Most people will want to use this specialized buffer pool.

```go
var buffPool = fsonutil.NewPool()

func Better() {
	buf := buffPool.Get()
	defer buffPool.Put(buf)
	
	obj := fson.NewObject(buf.Bytes())
	b := obj.String("hello", "world") // do things
    //...
}
```

This avoids allocating on the heap if the pool already contains a buffer.

### Reuse fson.Object

If you need to write out multiple JSON objects in the course of a single function you can reuse the same buffer. But!
You have to make sure to write out the result somewhere between each reuse.

```go
func Reuse() {
    // ...
	buf := buffPool.Get()
	defer buffPool.Put(buf)

    obj := fson.NewObject(buf.Bytes())

    first := obj.String("first", "message").Build()
    // Write out the response to STDOUT, after this you can safely reuse the buffer.
    fmt.Println(string(first)) // {"first":"message"}

    // Reset the internal buffer, ready for reuse 
    obj.Reset()

    // Use the same object
    second := obj.String("second", "message").Build()
    // Print out the second message
    fmt.Println(string(second)) // {"second":"message"}

    // ...
}
```

**NOTE**: This only works if you write out the result to some writer between reuses. `fson` will reset the buffer on 
a call to Reset() and will override the values that were already written to the slice. So this is wrong:


```go
func BAD_DO_NOT_DO_THIS() {
	// ...
	buf := buffPool.Get()
	defer buffPool.Put(buf)
	
	obj := fson.NewObject(buf.Bytes())
	first := obj.String("first", "message").Build()
	// THIS WILL OVERRIDE THE BUFFER but you have not written out the result yet!
	second := obj.String("second", "message").Build()
    
	fmt.Println(string(first))
	fmt.Println(string(second))
	
	// ...
}
```

## Benchmarks

Benchmarks are notoriously easy to manipulate and can be misleading but everybody wants to see the numbers so here they
are. These benchmarks are definitely "manipulated" to some degree. The most obvious tweak I made is to allocate a big
enough buffer for each benchmark so that the underlying append calls would never have to reallocate a new array.
Although to some degree they are fair because the stdlib tests get the same exact buffer size.

You can run the benchmarks yourself by running `make benchmark`. Running these on a Apple M3 Pro gives you roughly these
results.

```text
Benchmark                      Time/Op         Allocs/Op       Bytes/Op        vs Standard     Improvement    
-----------------------------  --------------- --------------- --------------- --------------- ---------------
BenchmarkObject_BuildSimple    16.73 ns/op     0 allocs/op     0 B/op          1x (baseline)   -              
BenchmarkJson_StdlibSimple     72.14 ns/op     96 allocs/op    1 B/op          4.31x           76.81%         
-----------------------------  --------------- --------------- --------------- --------------- ---------------
BenchmarkObject_BuildComplex   159.7 ns/op     0 allocs/op     0 B/op          1x (baseline)   -              
BenchmarkJson_StdlibComplex    281.4 ns/op     96 allocs/op    1 B/op          1.76x           43.25%         
-----------------------------  --------------- --------------- --------------- --------------- ---------------
BenchmarkObject_BuildLarge     17171 ns/op     0 allocs/op     0 B/op          1x (baseline)   -              
BenchmarkJson_StdlibLarge      22541 ns/op     630 allocs/op   12 B/op         1.31x           23.82%         
-----------------------------  --------------- --------------- --------------- --------------- ---------------

Summary                        fson            stdlib          Improvement    
-----------------------------  --------------- --------------- ---------------
Simple Case                    16.73 ns/op     72.14 ns/op     76.81%         
Complex Case                   159.7 ns/op     281.4 ns/op     43.25%         
Large Case                     17171 ns/op     22541 ns/op     23.82%         
Average Improvement            -               -               47.96% 
```


## A note on floating point values

Floats will be rendered as their numeric value. But there are some special values that might require some extra care.

Special float values like NaN and Infinity will be encoded as string values rather than JSON numbers, as JSON does not
support these values as numbers.

This means that arrays containing these special values will contain a mix of numeric types and string types. According
to RFC 8259 Section 5 (https://datatracker.ietf.org/doc/html/rfc8259#section-5) this is still valid JSON.

While this mixed-type array is valid JSON, it may cause issues when deserializing into strictly typed arrays. If you
need consistent types for deserialization, consider using the more explicit Start and Value API to control
which values are added to your array.

```go
package main

import (
	"fmt"
	"math"
	"github.com/LucasRouckhout/fson"
)

func main() {
	buf := make([]byte, 0, 1024)
	
	// Create values with some special floats
	values := []float64{
		1.23,
		math.NaN(),
		4.56,
		math.Inf(1),
		7.89,
		math.Inf(-1),
	}
	
	obj := fson.NewObject(buf).
        // This will replace the special float values with strings
        Floats64("withSpecialValues", values)
	    
    obj.Array("filteredValues")	
	for _, v := range values {
		// Filter out special values
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			obj.Float64Value(v)
		}
	}
	obj.EndArray()
	
	json := obj.Build()
	
	fmt.Println(string(json))
	// Output: {"withSpecialValues":[1.23,"NaN",4.56,"+Inf",7.89,"-Inf"],"filteredValues":[1.23,4.56,7.89]}
}
```






