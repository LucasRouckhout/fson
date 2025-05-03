# fson - High-performance JSON Encoder for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/LucasRouckhout/fson)](https://goreportcard.com/report/github.com/LucasRouckhout/fson)
[![GoDoc](https://godoc.org/github.com/LucasRouckhout/fson?status.svg)](https://godoc.org/github.com/LucasRouckhout/fson)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

`fson` is a high-performance JSON encoder for Go that focuses on simplicity and avoiding heap allocations. `fson` does one
thing and does it well: encode JSON. It explicitly does not try to solve other problems in the same space. If you need 
a serialization library `fson` is probably not the tool for you. The whole library is implemented as a single file 
of only ~1000LOC which makes it easy to validate and vendor if desired.

- **Fluent API**: Simple chainable interface for building JSON structures
- **Complete Control**: Full control over the produced JSON and heap allocations.
- **UTF8** `fson` guarantees to only produce valid UTF8 byte sequences.
- **No Reflection**: `fson` avoids reflection completely
- **Simple Implementation**: The entire library is contained in a single file of ~1000 lines (mostly documentation)
- **Easy to Vendor**: The small codebase makes it easy to vendor `fson` into an existing codebase

The finer details (especially UTF8 handling) of this library are heavily inspired by the json encoder of Uber's Zap logging library
[zapcore](https://github.com/uber-go/zap/tree/master/zapcore).

## Quick Example

```go
package main

import (
	"fmt"
	"github.com/LucasRouckhout/fson"
	"sync"
)

var buffPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024)
	},
}

func main() {
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).
		String("hello", "world").
		Bool("bool", true).
		Object("obj").
		String("foo", "bar").
		Bool("bool", false).
		Int("int", 8).
		EndObject().
		Build()

	fmt.Println(string(b))
	// -> {"foo":"bar","bool":true,"obj":{"foo":"bar","bool":false,"int":8}}
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

The raison d'être for `fson` is to allow developers full control over both the produced JSON and heap allocations. That's
to say that while `fson` itself will never allocate any memory on the heap you can still accidentally do so. Here are
some tips and tricks to use `fson` in an efficient manner.


### Use a sync.Pool

Avoid using make inside a function that is called multiple times within the lifetime of your program.

```go
func IfYouCallThisFunctionALotThisIsBad() {
	// ...
	buff := make([]byte, 0, 1024)
	fson.NewObject(buff)
	// ...
}
```

This will allocate a new buffer on the heap everytime the function is called. Instead, 
use a [sync.Pool](https://pkg.go.dev/sync#Pool) for this use case.

```go
var buffPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024)
	},
}

func Better() {
	// ...
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)
	
	fson.NewObject(buff)
	// ...
}
```

This avoids allocating on the heap if the pool already contains a buffer.

### Reuse of buffers

If you need to write out multiple JSON objects in the course of a single function you can reuse the same buffer. But!
You have to make sure to write out the result somewhere between each reuse.

```go
func Reuse() {
	// ...
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	first := fson.NewObject(buf).String("first", "message").Build()
	// Write out the response to STDOUT, after this you can safely reuse the buffer.
	fmt.Println(string(first))

	// Use the same buffer
	second := fson.NewObject(buf).String("second", "message").Build()
	// Print out the second message
	fmt.Println(string(second))

	// ...
}
```

**NOTE**: This only works if you write out the result to some writer between reuses. `fson` will reset the buffer on 
creation of a new object and will override the values of the underlying array the byte slice is pointing to. So this is
wrong!


```go
func BAD_DO_NOT_DO_THIS() {
	// ...
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	first := fson.NewObject(buf).String("first", "message").Build()
	// THIS WILL OVERRIDE THE BUFFER but you have not written out the result of first
	second := fson.NewObject(buf).String("second", "message").Build()
    
	fmt.Println(string(first))
	fmt.Println(string(second))
	// This prints out broken garbage!

	// ...
}
```

### How to shoot yourself in the foot

Coming soon...

### Benchmarks

Benchmarks are notoriously easy to manipulate and can be misleading but everybody wants to see the numbers so here they
are. These initial benchmarks were mostly written as a guide to make sure fson out-performs stdlib. They are very 
much open for improvements. 

As already mentioned, your millage may vary because how you use `fson` has a big impact on the performance gains it 
provides.

You can run the benchmarks yourself by running `make benchmark`. Running these on a Apple M3 Pro gives you roughly these
results.

```text
Benchmark                      Time/Op         Allocs/Op       Bytes/Op        vs Standard     Improvement    
-----------------------------  --------------- --------------- --------------- --------------- ---------------
BenchmarkObject_BuildSimple    33.52 ns/op     24 allocs/op    1 B/op          1x (baseline)   -              
BenchmarkJson_StdlibSimple     87.69 ns/op     133 allocs/op   1 B/op          2.62x           61.77%         
-----------------------------  --------------- --------------- --------------- --------------- ---------------
BenchmarkObject_BuildComplex   220.5 ns/op     232 allocs/op   2 B/op          1x (baseline)   -              
BenchmarkJson_StdlibComplex    337.4 ns/op     394 allocs/op   1 B/op          1.53x           34.65%         
-----------------------------  --------------- --------------- --------------- --------------- ---------------
BenchmarkObject_BuildLarge     22497 ns/op     61576 allocs/op 13 B/op         1x (baseline)   -              
BenchmarkJson_StdlibLarge      34079 ns/op     45891 allocs/op 12 B/op         1.51x           33.99%         
-----------------------------  --------------- --------------- --------------- --------------- ---------------

Summary                        fson            stdlib          Improvement    
-----------------------------  --------------- --------------- ---------------
Simple Case                    33.52 ns/op     87.69 ns/op     61.77%         
Complex Case                   220.5 ns/op     337.4 ns/op     34.65%         
Large Case                     22497 ns/op     34079 ns/op     33.99%         
Average Improvement            -               -               43.47% 
```

# Examples

### Basic Usage

Creating a simple JSON object:

```go
package main

import (
	"fmt"
	"github.com/LucasRouckhout/fson"
)

func main() {
	// Pre-allocate a buffer with enough capacity
	buf := make([]byte, 0, 1024)
	
	// Create a JSON object
	json := fson.NewObject(buf).
		String("name", "John Doe").
		Int("age", 30).
		Bool("active", true).
		Build()
	
	fmt.Println(string(json))
	// Output: {"name":"John Doe","age":30,"active":true}
}
```

### Arrays

Create an array of objects:

```go
package main

import (
	"fmt"
	"github.com/LucasRouckhout/fson"
)

func main() {
	buf := make([]byte, 0, 1024)
	
	// Create a JSON object with an array of objects
	json := fson.NewObject(buf).
		Array("people").
			StartObject().
				String("name", "Alice").
				Int("age", 28).
				Bool("active", true).
			EndObject().
			StartObject().
				String("name", "Bob").
				Int("age", 32).
				Bool("active", false).
			EndObject().
			StartObject().
				String("name", "Charlie").
				Int("age", 25).
				Bool("active", true).
			EndObject().
		EndArray().
		Build()
	
	fmt.Println(string(json))
	// Output: {"people":[{"name":"Alice","age":28,"active":true},{"name":"Bob","age":32,"active":false},{"name":"Charlie","age":25,"active":true}]}
}
```

### Floating point values

Floats will be rendered as their numeric value. But there are some special values that
might require some extra care.

Special float values like NaN and Infinity will be encoded as string values
rather than JSON numbers, as JSON does not support these values as numbers.

This means that arrays containing these special values will contain a mix of
numeric types and string types. According to RFC 8259 Section 5
(https://datatracker.ietf.org/doc/html/rfc8259#section-5) this is still valid JSON.

While this mixed-type array is valid JSON, it may cause issues when
deserializing into strictly typed arrays. If you need consistent types for deserialization, consider using the more
explicit Start and Value API.

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

### Multi-Type Arrays

Creating arrays with mixed types:

```go
package main

import (
	"fmt"
	"time"
	"github.com/LucasRouckhout/fson"
)

func main() {
	buf := make([]byte, 0, 1024)

	// Create a JSON object with a multi-type array
	obj := fson.NewObject(buf).
		// Use the Array method to start a heterogeneous array
		Array("mixedTypes").
		StringValue("text value").
		IntValue(42).
		BoolValue(true).
		Float64Value(3.14159).
		// Add different types of values to the array
		StartObject().
		String("key", "value").
		EndObject().
		StartArray().
		IntValue(1).
		IntValue(2).
		EndArray().
		NullValue().
		EndArray()

	json := obj.Build()

	fmt.Println(string(json))
	// Output: {"mixedTypes":["text value",42,true,3.14159,{"key":"value"},[1,2],null]}
}

```

### Date and Time Values

Handling date and time values:

```go
package main

import (
	"fmt"
	"time"
	"github.com/LucasRouckhout/fson"
)

func main() {
	buf := make([]byte, 0, 1024)
	
	// Create some time values
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	dur := 1*time.Hour + 23*time.Minute + 45*time.Second

	// Create a JSON object with time values
	json := fson.NewObject(buf).
		Time("now", now, time.RFC3339).
		Time("iso8601", now, time.RFC3339).
		Time("rfc822", now, time.RFC822).
		Time("custom", now, "2006-01-02").
		Times("schedule", []time.Time{yesterday, now, tomorrow}, time.RFC3339).
		// Duration is stored as a human-readable string using the String() method
		Duration("elapsed", dur).
		// Use Int64 to store duration as nanoseconds (or any other unit) if needed
		Int64("elapsedNanos", dur.Nanoseconds()).
		Build()
	
	fmt.Println(string(json))
	// Output will contain formatted dates and times
}
```

### UTF8 and character escaping

Handling strings with special characters:

```go
package main

import (
	"fmt"
	"github.com/LucasRouckhout/fson"
)

func main() {
	buf := make([]byte, 0, 1024)
	
	// Create a JSON object with strings containing special characters
	json := fson.NewObject(buf).
		String("simpleText", "Hello World").
		String("quotedText", "She said, \"Hello!\"").
		String("escapedChars", "Tab:\t Newline:\n Backslash:\\").
		String("unicodeChars", "Emoji: 😊 Kanji: 漢字").
		String("controlChars", string([]byte{0x01, 0x02})).
		String("path", "C:\\Program Files\\App\\config.json").
		String("html", "<div>Some HTML content</div>").
		Build()
	
	fmt.Println(string(json))
	// All special characters will be properly escaped in the output
}
```

### Working with Null Values

JSON allows for explicit `null` values, which represent the absence of a value or an undefined value. In fson, you can easily work with null values using the dedicated null-handling functions.

The `Null()` function lets you quickly add a key with a null value:

```go
package main

import (
	"fmt"
	"github.com/LucasRouckhout/fson"
)

func main() {
	buf := make([]byte, 0, 256)
	
	json := fson.NewObject(buf).
		String("name", "John Doe").
		Int("age", 30).
		Null("address").    // Explicitly set address to null
		Build()
	
	fmt.Println(string(json))
	// Output: {"name":"John Doe","age":30,"address":null}
}
```

When using the explicit key-value approach, you can use NullValue() after a key or inside of an array.

```go
package main

import (
	"fmt"
	"github.com/LucasRouckhout/fson"
)

func main() {
	buf := make([]byte, 0, 256)
	
	json := fson.NewObject(buf).
		Array("items").
			StringValue("first").
			NullValue().          // Add a null element
			StringValue("third").
		EndArray().
		
		// You can also create arrays with mixed types including nulls
		Array("mixed").
			IntValue(42).
			NullValue().          // Add a null element
			BoolValue(true).
			StringValue("text").
		EndArray().
		Build()
	
	fmt.Println(string(json))
	// Output: {"items":["first",null,"third"],"mixed":[42,null,true,"text"]}
}
```

# A note on performance

Using `fson` won't make your code fast "out-of-the-box" but will enable you to
more efficiently encode JSON. Here are some tips and tricks to use `fson` efficiently.

### Make use of sync.Pool











