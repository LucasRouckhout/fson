# fson - Fast, Allocation-free JSON Encoder for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/LucasRouckhout/fson)](https://goreportcard.com/report/github.com/LucasRouckhout/fson)
[![GoDoc](https://godoc.org/github.com/LucasRouckhout/fson?status.svg)](https://godoc.org/github.com/LucasRouckhout/fson)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**fson** is a high-performance JSON encoder for Go that focuses on zero allocations, performance, and full control of
the generated json. 

`fson` does one thing and does it well: encoding Go data into JSON. This package is mainly aimed at people who need full
control over both the produced JSON and memory allocations.

## Features

- **Zero Allocations**: Works with pre-allocated buffers to minimize GC pressure. You are in control over memory allocations.
- **Fluent API**: Simple chainable interface for building JSON structures
- **Complete Control**: Handles all JSON data types with special care for edge cases
- **UTF-8 Support**: Properly handles and escapes all Unicode characters
- **No Reflection**: Direct encoding without using reflection
- **Simple Implementation**: The entire library is contained in a single file of ~1000 lines (mostly documentation)
- **Easy to Vendor**: Small codebase makes it trivial to vendor and customize for your specific needs

## Quick Example

```go
package main

import (
    "fmt"
    "github.com/LucasRouckhout/fson"
)

func main() {
    // Pre-allocate a buffer with enough capacity
    buf := make([]byte, 0, 1024)
    
    // Create a JSON object using the fluent API
    json := fson.NewObject(buf).
        String("name", "John Doe").
        Int("age", 30).
        Bool("active", true).
        Array("tags").
            StringValue("json").
            StringValue("performance").
            StringValue("go").
        EndArray().
        Build()
    
    fmt.Println(string(json))
    // Output: {"name":"John Doe","age":30,"active":true,"tags":["json","performance","go"]}
}
```

# Usage

Add `fson` as a dependency

```
go get -u github.com/LucasRouckhout/fson
```

## Examples

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

### Using Key-Value Methods

You can use either the combined methods or the separate Key/Value methods:

```go
package main

import (
	"fmt"
	"github.com/LucasRouckhout/fson"
)

func main() {
	buf := make([]byte, 0, 1024)
	
	// Using combined methods
	json1 := fson.NewObject(buf).
		String("name", "John Doe").
		Int("age", 30).
		Build()
	
	// Reset buffer
	buf = buf[:0]
	
	// Using separate Key/Value methods - gives more flexibility
	json2 := fson.NewObject(buf).
		Key("name").StringValue("John Doe").
		Key("age").IntValue(30).
		Build()
	
	fmt.Println(string(json1))
	fmt.Println(string(json2))
	// Both produce: {"name":"John Doe","age":30}
}
```

### Nested Objects

Working with nested objects:

```go
package main

import (
	"fmt"
	"github.com/LucasRouckhout/fson"
)

func main() {
	buf := make([]byte, 0, 1024)
	
	// Create a JSON object with nested objects
	json := fson.NewObject(buf).
		String("name", "John Doe").
		Int("age", 30).
		Object("address").
			String("street", "123 Main St").
			String("city", "Anytown").
			String("country", "USA").
			Int("zipCode", 12345).
		EndObject().
		Object("contact").
			String("email", "john@example.com").
			String("phone", "+1234567890").
		EndObject().
		Build()
	
	fmt.Println(string(json))
	// Output: {"name":"John Doe","age":30,"address":{"street":"123 Main St","city":"Anytown","country":"USA","zipCode":12345},"contact":{"email":"john@example.com","phone":"+1234567890"}}
}
```

### Working with Arrays

Simple arrays of primitive values:

```go
package main

import (
	"fmt"
	"github.com/LucasRouckhout/fson"
)

func main() {
	buf := make([]byte, 0, 1024)
	
	// Create a JSON object with arrays
	json := fson.NewObject(buf).
		String("name", "John Doe").
		Ints("scores", []int{85, 90, 92, 88}).
		Strings("hobbies", []string{"reading", "hiking", "photography"}).
		Bools("settings", []bool{true, false, true}).
		Build()
	
	fmt.Println(string(json))
	// Output: {"name":"John Doe","scores":[85,90,92,88],"hobbies":["reading","hiking","photography"],"settings":[true,false,true]}
}
```

### Array of Objects

Creating arrays of objects:

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

### Working with Floating-Point Values

Handling floating-point values. Special values like NaN and Infinity will be encoded as string values. If you need strict
typing for float arrays take a look at the section about [Filtering Nan and Infinity Values](Filtering%20Nan%20and%20Infinity%20Values)

```go
package main

import (
	"fmt"
	"math"
	"github.com/LucasRouckhout/fson"
)

func main() {
	buf := make([]byte, 0, 1024)
	
	// Create a JSON object with regular and special float values
	json := fson.NewObject(buf).
		Float64("regularValue", 3.14159).
		Float64("nanValue", math.NaN()).
		Float64("positiveInfinity", math.Inf(1)).
		Float64("negativeInfinity", math.Inf(-1)).
		
		// Array with mixed float values
		Floats64("mixedFloats", []float64{
			1.23,
			math.NaN(),
			4.56,
			math.Inf(1),
			7.89,
		}).
		Build()
	
	fmt.Println(string(json))
	// Output example: {"regularValue":3.14159,"nanValue":"NaN","positiveInfinity":"+Inf","negativeInfinity":"-Inf","mixedFloats":[1.23,"NaN",4.56,"+Inf",7.89]}
}
```

### Filtering Nan and Infinity Values

Special values like NaN and Infinity will be encoded as string values
rather than JSON numbers, as JSON does not support these values as numbers.

This means that arrays containing these special values will contain a mix of
numeric types and string types. According to RFC 8259 Section 5
(https://datatracker.ietf.org/doc/html/rfc8259#section-5) this is still valid JSON:
"There is no requirement that the values in an array be of the same type."

While this mixed-type array is valid JSON, it may cause issues when
deserializing into strictly typed arrays. If you need consistent types for deserialization, consider using the more
explicit StartArray() approach and handling special values manually:

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
	
	// Create a JSON object with filtered floats
	obj := fson.NewObject(buf)
	
	// Standard approach (includes special values as strings)
	obj.Floats64("withSpecialValues", values)
	
	// Custom approach (filters out special values)
	obj.Key("filteredValues").StartArray()
	for _, v := range values {
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
	obj := fson.NewObject(buf)
	
	// Use the Array method to start a heterogeneous array
	obj.Array("mixedTypes")
	
	// Add different types of values to the array
	obj.StringValue("text value")
	obj.IntValue(42)
	obj.BoolValue(true)
	obj.Float64Value(3.14159)
	obj.StartObject()
		obj.String("key", "value")
	obj.EndObject()
	obj.StartArray()
		obj.IntValue(1)
		obj.IntValue(2)
	obj.EndArray()
	obj.StringValue(nil) // Add null value
	
	// End the array and build the JSON
	obj.EndArray()
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
	
	// Create a JSON object with time values
	json := fson.NewObject(buf).
		Time("now", now, time.RFC3339).
		Time("iso8601", now, time.RFC3339).
		Time("rfc822", now, time.RFC822).
		Time("custom", now, "2006-01-02").
		Times("schedule", []time.Time{yesterday, now, tomorrow}, time.RFC3339).
		// Duration is stored as a human-readable string
		Duration("elapsed", 1*time.Hour + 23*time.Minute + 45*time.Second).
		// Use Int64 to store duration as nanoseconds if needed
		Int64("elapsedNanos", (1*time.Hour + 23*time.Minute + 45*time.Second).Nanoseconds()).
		Build()
	
	fmt.Println(string(json))
	// Output will contain formatted dates and times
}
```

### Null values


### Custom String Formatting and Escaping

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








