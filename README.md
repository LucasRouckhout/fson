# Fson

Simple (less than 500 LOC), Performant and allocation-free JSON encoder. Heavily inspired by the encoders in both
zerolog and Uber's Zap logging packages. 

`fson` exposes a fluent-like API to build up valid JSON byte representations.

```golang
package main

import (
	"fmt"
	"github.com/LucasRouckhout/fson"
	"sync"
)

var buffPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 100)
	},
}

func main() {
	// Get yourself a buffer
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).
		String("hello", "world").
		Bool("bool", true).
		StartObject("obj").
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

`fson` is ideal for people who want to permanently produce valid JSON representations of their structs.

