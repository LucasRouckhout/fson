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
