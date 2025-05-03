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
		Object("obj").
		String("foo", "bar").
		Bool("bool", false).
		Int("int", 8).
		EndObject().
		Build()

	fmt.Println(string(b)) // -> {"foo":"bar","bool":true,"obj":{"foo":"bar","bool":false,"int":8}}
}
