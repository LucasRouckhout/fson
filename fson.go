// Copyright (c) 2025 Lucas Rocukhout
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package fson exposed a high-performance JSON encoder that focuses on simplicity
// and avoiding heap allocations.
//
// # Usage
//
// `fson` exposes a "two-tiered" fluent API that enables method chaining to build up a JSON object. For example
// this code snippet:
//
//	fson.NewObject(buf).Key("foo").StringValue("bar").Build()
//
// Would produce:
//
//	{"foo": "bar"}
//
// Or a more complex example
//
//	fson.NewObject(buf).Key("fooArr").StartArray().StringValue("bar").StringValue("foo").EndArray().Build()
//
// Would produce:
//
//	{"fooArr":["bar","foo"]}
//
// This explicit key-value approach can become cumbersome and verbose. So for most operations there exists a shorthand
// "higher-level" API. These two examples can respectively be rewritten in a shorter manner like this.
//
//	fson.NewObject(buf).String("foo", "bar").Build()
//	fson.NewObject(buf).Strings("fooArr", []string{"bar", "foo"}).Build()
//
// For most use-cases the higher-level API will be enough. But there are examples, like multi-typed arrays, where you will
// need to fall back to the lower level API to produce the desired output.
//
// # A note on performance
//
// The raison d'être for `fson` is to allow developers full control over both the produced JSON and heap allocations. That's
// to say that while `fson` itself will never allocate any memory on the heap you can still accidentally do so. Here are
// some tips and tricks to use `fson` in an efficient manner.
//
// Avoid using make inside a function that is called multiple times within the lifetime of your program.
//
//	func IfYouCallThisFunctionALotThisIsBad() {
//		// ...
//		buff := make([]byte, 0, 1024)
//		fson.NewObject(buff)
//		// ...
//	}
//
// This will allocate a new buffer on the heap everytime the function is called. Instead,
// use a [sync.Pool](https://pkg.go.dev/sync#Pool) for this use case.
//
//	var buffPool = sync.Pool{
//		New: func() interface{} {
//			return make([]byte, 0, 1024)
//		},
//	}
//
//	func Better() {
//		// ...
//		buf := buffPool.Get().([]byte)
//		defer buffPool.Put(buf)
//
//		fson.NewObject(buff)
//		// ...
//	}
//
// This avoids allocating on the heap if the pool already contains a buffer.
//
// If you need to write out multiple JSON objects in the course of a single function you can reuse the same buffer. But!
// You have to make sure to write out the result somewhere between each reuse.
//
//	func Reuse() {
//		// ...
//		buf := buffPool.Get().([]byte)
//		defer buffPool.Put(buf)
//
//		first := fson.NewObject(buf).String("first", "message").Build()
//		// Write out the response to STDOUT, after this you can safely reuse the buffer.
//		fmt.Println(string(first))
//
//		// Use the same buffer
//		second := fson.NewObject(buf).String("second", "message").Build()
//		// Print out the second message
//		fmt.Println(string(second))
//
//		// ...
//	}
//
// **NOTE**: This only works if you write out the result to some writer between reuses. `fson` will reset the buffer on
// creation of a new object and will override the values of the underlying array the byte slice is pointing to. So this is
// wrong!
//
//	func BAD_DO_NOT_DO_THIS() {
//		// ...
//		buf := buffPool.Get().([]byte)
//		defer buffPool.Put(buf)
//
//		first := fson.NewObject(buf).String("first", "message").Build()
//		// THIS WILL OVERRIDE THE BUFFER but you have not written out the result of first
//		second := fson.NewObject(buf).String("second", "message").Build()
//
//		fmt.Println(string(first))
//		fmt.Println(string(second))
//		// This prints out broken garbage!
//
//		// ...
//	}
package fson

import (
	"math"
	"strconv"
	"time"
	"unicode/utf8"
)

// Object represents a JSON object being constructed.
// It maintains an internal byte buffer where the JSON is incrementally built up.
type Object struct {
	buf []byte
}

// NewObject creates a new JSON object builder using the provided byte buffer.
// NewObject will reset the provided buffer before use.
//
// The caller is responsible for ensuring the buffer has sufficient capacity
// to hold the complete JSON structure. If the buffer is too small, append
// operations may cause reallocations, reducing performance benefits.
func NewObject(buf []byte) *Object {
	obj := &Object{
		buf[:0], // Reset buffer
	}

	obj.buf = append(obj.buf, '{')

	return obj
}

// Key appends a key to the JSON object and prepares for a value to be added.
//
// Note that calling Key() without a subsequent Value method call will result in
// incomplete and invalid JSON. Always follow Key() with an appropriate Value method.
//
// Example:
//
//	obj.Key("name").StringValue("John")
//	// Results in: {"name":"John"}
//
// The Key function is part of the low-level API that gives more control
// over JSON construction compared to the combined methods. After calling Key(),
// you should call one of the Value methods (StringValue, IntValue, etc.) to add
// the corresponding value for this key.
func (o *Object) Key(key string) *Object {
	o.buf = appendString(o.buf, key)
	o.buf = append(o.buf, ':')
	return o
}

// Null appends a null value with the specified key to the JSON object.
// This creates a key-value pair where the value is explicitly set to JSON null.
//
// Example:
//
//	obj.Null("optionalField")
//	// Results in: {"optionalField":null}
//
// This method is useful for explicitly representing missing or undefined values
// according to the JSON specification.
func (o *Object) Null(key string) *Object {
	o.Key(key).NullValue()
	return o
}

// NullValue appends a null value to the current key in the JSON object.
//
// Example:
//
//	obj.Key("optionalField").NullValue()
//	// Results in: {"optionalField":null}
//
// This method should be used after calling Key() when you want to explicitly
// set a value to null rather than omitting the field entirely.
func (o *Object) NullValue() *Object {
	o.buf = append(o.buf, "null"...)
	o.buf = append(o.buf, ',')
	return o
}

// String appends a string key-value pair to the JSON object.
//
// Example:
//
//	obj.String("name", "John Doe")
func (o *Object) String(key, value string) *Object {
	return o.Key(key).StringValue(value)
}

// StringValue appends a string value to the current key in the JSON object.
//
// Example:
//
//	obj.Key("name").StringValue("John Doe")
func (o *Object) StringValue(value string) *Object {
	o.buf = appendString(o.buf, value)
	o.buf = append(o.buf, ',')
	return o
}

// Strings appends an array of strings as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Strings("tags", []string{"json", "encoder", "go"})
func (o *Object) Strings(key string, value []string) *Object {
	return o.Key(key).StringsValue(value)
}

// StringsValue appends an array of strings to the current key in the JSON object.
//
// Example:
//
//	obj.Key("tags").StringsValue([]string{"json", "encoder", "go"})
func (o *Object) StringsValue(value []string) *Object {
	o.buf = appendArray(o.buf, value, appendString)
	o.buf = append(o.buf, ',')
	return o
}

// Int appends an integer key-value pair to the JSON object.
// This is a convenience wrapper around Int64.
//
// Example:
//
//	obj.Int("count", 42)
func (o *Object) Int(key string, value int) *Object {
	return o.Key(key).IntValue(value)
}

// IntValue appends an integer value to the current key in the JSON object.
// This is a convenience wrapper around Int64Value.
//
// Example:
//
//	obj.Key("count").IntValue(42)
func (o *Object) IntValue(value int) *Object {
	return o.Int64Value(int64(value))
}

// Ints appends an array of integers as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Ints("values", []int{1, 2, 3, 4, 5})
func (o *Object) Ints(key string, value []int) *Object {
	return o.Key(key).IntsValue(value)
}

// IntsValue appends an array of integers to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").IntsValue([]int{1, 2, 3, 4, 5})
func (o *Object) IntsValue(value []int) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value int) []byte {
		return strconv.AppendInt(buf, int64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Int8 appends an int8 key-value pair to the JSON object.
// This is a convenience wrapper around Int64.
//
// Example:
//
//	obj.Int8("value", 42)
func (o *Object) Int8(key string, value int8) *Object {
	return o.Key(key).Int8Value(value)
}

// Int8Value appends an int8 value to the current key in the JSON object.
// This is a convenience wrapper around Int64Value.
//
// Example:
//
//	obj.Key("value").Int8Value(42)
func (o *Object) Int8Value(value int8) *Object {
	return o.Int64Value(int64(value))
}

// Ints8 appends an array of int8 values as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Ints8("values", []int8{1, 2, 3, 4, 5})
func (o *Object) Ints8(key string, value []int8) *Object {
	return o.Key(key).Ints8Value(value)
}

// Ints8Value appends an array of int8 values to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").Ints8Value([]int8{1, 2, 3, 4, 5})
func (o *Object) Ints8Value(value []int8) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value int8) []byte {
		return strconv.AppendInt(buf, int64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Int16 appends an int16 key-value pair to the JSON object.
// This is a convenience wrapper around Int64.
//
// Example:
//
//	obj.Int16("value", 42)
func (o *Object) Int16(key string, value int16) *Object {
	return o.Key(key).Int16Value(value)
}

// Int16Value appends an int16 value to the current key in the JSON object.
// This is a convenience wrapper around Int64Value.
//
// Example:
//
//	obj.Key("value").Int16Value(42)
func (o *Object) Int16Value(value int16) *Object {
	return o.Int64Value(int64(value))
}

// Ints16 appends an array of int16 values as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Ints16("values", []int16{1, 2, 3, 4, 5})
func (o *Object) Ints16(key string, value []int16) *Object {
	return o.Key(key).Ints16Value(value)
}

// Ints16Value appends an array of int16 values to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").Ints16Value([]int16{1, 2, 3, 4, 5})
func (o *Object) Ints16Value(value []int16) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value int16) []byte {
		return strconv.AppendInt(buf, int64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Int32 appends an int32 key-value pair to the JSON object.
// This is a convenience wrapper around Int64.
//
// Example:
//
//	obj.Int32("value", 42)
func (o *Object) Int32(key string, value int32) *Object {
	return o.Key(key).Int32Value(value)
}

// Int32Value appends an int32 value to the current key in the JSON object.
// This is a convenience wrapper around Int64Value.
//
// Example:
//
//	obj.Key("value").Int32Value(42)
func (o *Object) Int32Value(value int32) *Object {
	return o.Int64Value(int64(value))
}

// Ints32 appends an array of int32 values as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Ints32("values", []int32{1, 2, 3, 4, 5})
func (o *Object) Ints32(key string, value []int32) *Object {
	return o.Key(key).Ints32Value(value)
}

// Ints32Value appends an array of int32 values to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").Ints32Value([]int32{1, 2, 3, 4, 5})
func (o *Object) Ints32Value(value []int32) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value int32) []byte {
		return strconv.AppendInt(buf, int64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Int64 appends an int64 key-value pair to the JSON object.
// This is the base method that other integer methods call internally.
//
// Example:
//
//	obj.Int64("value", 42)
func (o *Object) Int64(key string, value int64) *Object {
	return o.Key(key).Int64Value(value)
}

// Int64Value appends an int64 value to the current key in the JSON object.
// This is the base method that other integer value methods call internally.
//
// Example:
//
//	obj.Key("value").Int64Value(42)
func (o *Object) Int64Value(value int64) *Object {
	o.buf = strconv.AppendInt(o.buf, value, 10)
	o.buf = append(o.buf, ',')
	return o
}

// Ints64 appends an array of int64 values as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Ints64("values", []int64{1, 2, 3, 4, 5})
func (o *Object) Ints64(key string, value []int64) *Object {
	return o.Key(key).Ints64Value(value)
}

// Ints64Value appends an array of int64 values to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").Ints64Value([]int64{1, 2, 3, 4, 5})
func (o *Object) Ints64Value(value []int64) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value int64) []byte {
		return strconv.AppendInt(buf, value, 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Uint appends an unsigned integer key-value pair to the JSON object.
// This is a convenience wrapper around Uint64.
//
// Example:
//
//	obj.Uint("count", 42)
func (o *Object) Uint(key string, value uint) *Object {
	return o.Key(key).UintValue(value)
}

// UintValue appends an unsigned integer value to the current key in the JSON object.
// This is a convenience wrapper around Uint64Value.
//
// Example:
//
//	obj.Key("count").UintValue(42)
func (o *Object) UintValue(value uint) *Object {
	return o.Uint64Value(uint64(value))
}

// Uints appends an array of unsigned integers as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Uints("values", []uint{1, 2, 3, 4, 5})
func (o *Object) Uints(key string, value []uint) *Object {
	return o.Key(key).UintsValue(value)
}

// UintsValue appends an array of unsigned integers to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").UintsValue([]uint{1, 2, 3, 4, 5})
func (o *Object) UintsValue(value []uint) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value uint) []byte {
		return strconv.AppendUint(buf, uint64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Uint8 appends a uint8 key-value pair to the JSON object.
// This is a convenience wrapper around Uint64.
//
// Example:
//
//	obj.Uint8("value", 42)
func (o *Object) Uint8(key string, value uint8) *Object {
	return o.Key(key).Uint8Value(value)
}

// Uint8Value appends a uint8 value to the current key in the JSON object.
// This is a convenience wrapper around Uint64Value.
//
// Example:
//
//	obj.Key("value").Uint8Value(42)
func (o *Object) Uint8Value(value uint8) *Object {
	return o.Uint64Value(uint64(value))
}

// Uints8 appends an array of uint8 values as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Uints8("values", []uint8{1, 2, 3, 4, 5})
func (o *Object) Uints8(key string, value []uint8) *Object {
	return o.Key(key).Uints8Value(value)
}

// Uints8Value appends an array of uint8 values to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").Uints8Value([]uint8{1, 2, 3, 4, 5})
func (o *Object) Uints8Value(value []uint8) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value uint8) []byte {
		return strconv.AppendUint(buf, uint64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Uint16 appends a uint16 key-value pair to the JSON object.
// This is a convenience wrapper around Uint64.
//
// Example:
//
//	obj.Uint16("value", 42)
func (o *Object) Uint16(key string, value uint16) *Object {
	return o.Key(key).Uint16Value(value)
}

// Uint16Value appends a uint16 value to the current key in the JSON object.
// This is a convenience wrapper around Uint64Value.
//
// Example:
//
//	obj.Key("value").Uint16Value(42)
func (o *Object) Uint16Value(value uint16) *Object {
	return o.Uint64Value(uint64(value))
}

// Uints16 appends an array of uint16 values as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Uints16("values", []uint16{1, 2, 3, 4, 5})
func (o *Object) Uints16(key string, value []uint16) *Object {
	return o.Key(key).Uints16Value(value)
}

// Uints16Value appends an array of uint16 values to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").Uints16Value([]uint16{1, 2, 3, 4, 5})
func (o *Object) Uints16Value(value []uint16) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value uint16) []byte {
		return strconv.AppendUint(buf, uint64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Uint32 appends a uint32 key-value pair to the JSON object.
// This is a convenience wrapper around Uint64.
//
// Example:
//
//	obj.Uint32("value", 42)
func (o *Object) Uint32(key string, value uint32) *Object {
	return o.Key(key).Uint32Value(value)
}

// Uint32Value appends a uint32 value to the current key in the JSON object.
// This is a convenience wrapper around Uint64Value.
//
// Example:
//
//	obj.Key("value").Uint32Value(42)
func (o *Object) Uint32Value(value uint32) *Object {
	return o.Uint64Value(uint64(value))
}

// Uints32 appends an array of uint32 values as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Uints32("values", []uint32{1, 2, 3, 4, 5})
func (o *Object) Uints32(key string, value []uint32) *Object {
	return o.Key(key).Uints32Value(value)
}

// Uints32Value appends an array of uint32 values to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").Uints32Value([]uint32{1, 2, 3, 4, 5})
func (o *Object) Uints32Value(value []uint32) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value uint32) []byte {
		return strconv.AppendUint(buf, uint64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Uint64 appends a uint64 key-value pair to the JSON object.
// This is the base method that other unsigned integer methods call internally.
//
// Example:
//
//	obj.Uint64("value", 42)
func (o *Object) Uint64(key string, value uint64) *Object {
	return o.Key(key).Uint64Value(value)
}

// Uint64Value appends a uint64 value to the current key in the JSON object.
// This is the base method that other unsigned integer value methods call internally.
//
// Example:
//
//	obj.Key("value").Uint64Value(42)
func (o *Object) Uint64Value(value uint64) *Object {
	o.buf = strconv.AppendUint(o.buf, value, 10)
	o.buf = append(o.buf, ',')
	return o
}

// Uints64 appends an array of uint64 values as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Uints64("values", []uint64{1, 2, 3, 4, 5})
func (o *Object) Uints64(key string, value []uint64) *Object {
	return o.Key(key).Uints64Value(value)
}

// Uints64Value appends an array of uint64 values to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").Uints64Value([]uint64{1, 2, 3, 4, 5})
func (o *Object) Uints64Value(value []uint64) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value uint64) []byte {
		return strconv.AppendUint(buf, value, 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Float32 appends a float32 key-value pair to the JSON object.
// This is a convenience wrapper around Float64.
//
// Example:
//
//	obj.Float32("value", 3.14)
//
// Note: Special values like NaN and Infinity will be encoded as string values
// rather than JSON numbers, as JSON does not support these values as numbers.
func (o *Object) Float32(key string, value float32) *Object {
	return o.Key(key).Float32Value(value)
}

// Float32Value appends a float32 value to the current key in the JSON object.
// This is a convenience wrapper around Float64Value.
//
// Example:
//
//	obj.Key("value").Float32Value(3.14)
//
// Note: Special values like NaN and Infinity will be encoded as string values
// rather than JSON numbers, as JSON does not support these values as numbers.
func (o *Object) Float32Value(value float32) *Object {
	return o.Float64Value(float64(value))
}

// Floats32 appends an array of float32 values as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Floats32("values", []float32{1.1, 2.2, 3.3})
//
// Note: Special values like NaN and Infinity will be encoded as string values
// rather than JSON numbers, as JSON does not support these values as numbers.
//
// This means that arrays containing these special values will contain a mix of
// numeric types and string types. According to RFC 8259 Section 5
// (https://datatracker.ietf.org/doc/html/rfc8259#section-5):
// "There is no requirement that the values in an array be of the same type."
//
// While this mixed-type array is valid JSON, it may cause issues when
// deserializing into strictly typed arrays.
// If you need consistent types for deserialization, consider using the more
// explicit StartArray() approach and handling special values manually:
//
//	obj.Key("values").StartArray()
//	for _, v := range floatValues {
//	    if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
//	        // Handle special values differently or skip them
//	        continue
//	    }
//	    obj.Float32Value(v)
//	}
//	obj.EndArray()
func (o *Object) Floats32(key string, value []float32) *Object {
	return o.Key(key).Floats32Value(value)
}

// Floats32Value appends an array of float32 values to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").Floats32Value([]float32{1.1, 2.2, 3.3})
//
// Note: Special values like NaN and Infinity will be encoded as string values
// rather than JSON numbers, as JSON does not support these values as numbers.
func (o *Object) Floats32Value(value []float32) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value float32) []byte {
		return appendFloat(buf, float64(value), 32)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Float64 appends a float64 key-value pair to the JSON object.
// This is the base method that other floating-point methods call internally.
//
// Example:
//
//	obj.Float64("value", 3.14159265359)
//
// Note: Special values like NaN and Infinity will be encoded as string values
// rather than JSON numbers, as JSON does not support these values as numbers.
func (o *Object) Float64(key string, value float64) *Object {
	return o.Key(key).Float64Value(value)
}

// Float64Value appends a float64 value to the current key in the JSON object.
// This is the base method that other floating-point value methods call internally.
//
// Example:
//
//	obj.Key("value").Float64Value(3.14159265359)
//
// Note: Special values like NaN and Infinity will be encoded as string values
// rather than JSON numbers, as JSON does not support these values as numbers.
func (o *Object) Float64Value(value float64) *Object {
	o.buf = appendFloat(o.buf, value, 64)
	o.buf = append(o.buf, ',')
	return o
}

// Floats64 appends an array of float64 values as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Floats64("values", []float64{1.1, 2.2, 3.3})
//
// Note: Special values like NaN and Infinity will be encoded as string values
// rather than JSON numbers, as JSON does not support these values as numbers.
//
// This means that arrays containing these special values will contain a mix of
// numeric types and string types. According to RFC 8259 Section 5
// (https://datatracker.ietf.org/doc/html/rfc8259#section-5):
// "There is no requirement that the values in an array be of the same type."
//
// While this mixed-type array is valid JSON, it may cause issues when
// deserializing into strictly typed arrays.
// If you need consistent types for deserialization, consider using the more
// explicit StartArray() approach and handling special values manually:
//
//	obj.Key("values").StartArray()
//	for _, v := range floatValues {
//	    if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
//	        // Handle special values differently or skip them
//	        continue
//	    }
//	    obj.Float64Value(v)
//	}
//	obj.EndArray()
func (o *Object) Floats64(key string, value []float64) *Object {
	return o.Key(key).Floats64Value(value)
}

// Floats64Value appends an array of float64 values to the current key in the JSON object.
//
// Example:
//
//	obj.Key("values").Floats64Value([]float64{1.1, 2.2, 3.3})
//
// Note: Special values like NaN and Infinity will be encoded as string values
// rather than JSON numbers, as JSON does not support these values as numbers.
func (o *Object) Floats64Value(value []float64) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value float64) []byte {
		return appendFloat(buf, value, 64)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Bool appends a boolean key-value pair to the JSON object.
//
// Example:
//
//	obj.Bool("active", true)
func (o *Object) Bool(key string, value bool) *Object {
	return o.Key(key).BoolValue(value)
}

// BoolValue appends a boolean value to the current key in the JSON object.
//
// Example:
//
//	obj.Key("active").BoolValue(true)
func (o *Object) BoolValue(value bool) *Object {
	o.buf = strconv.AppendBool(o.buf, value)
	o.buf = append(o.buf, ',')
	return o
}

// Bools appends an array of boolean values as a key-value pair to the JSON object.
//
// Example:
//
//	obj.Bools("flags", []bool{true, false, true})
func (o *Object) Bools(key string, value []bool) *Object {
	return o.Key(key).BoolsValue(value)
}

// BoolsValue appends an array of boolean values to the current key in the JSON object.
//
// Example:
//
//	obj.Key("flags").BoolsValue([]bool{true, false, true})
func (o *Object) BoolsValue(value []bool) *Object {
	o.buf = appendArray(o.buf, value, strconv.AppendBool)
	o.buf = append(o.buf, ',')
	return o
}

// Time appends a time.Time key-value pair to the JSON object.
// The time is formatted as a string according to the specified format.
//
// Example:
//
//	obj.Time("created", time.Now(), time.RFC3339)
//
// The time will be encoded as a JSON string value with proper quotation marks.
// Common formats include time.RFC3339, time.RFC822, and time.RFC1123.
func (o *Object) Time(key string, value time.Time, format string) *Object {
	return o.Key(key).TimeValue(value, format)
}

// TimeValue appends a time.Time value to the current key in the JSON object.
// The time is formatted as a string according to the specified format.
//
// Example:
//
//	obj.Key("created").TimeValue(time.Now(), time.RFC3339)
//
// The time will be encoded as a JSON string value with proper quotation marks.
// Common formats include time.RFC3339, time.RFC822, and time.RFC1123.
func (o *Object) TimeValue(value time.Time, format string) *Object {
	o.buf = appendTime(o.buf, value, format)
	o.buf = append(o.buf, ',')
	return o
}

// Times appends an array of time.Time values as a key-value pair to the JSON object.
// All times are formatted as strings according to the specified format.
//
// Example:
//
//	obj.Times("timestamps", []time.Time{time.Now(), time.Now().Add(-24*time.Hour)}, time.RFC3339)
//
// Each time will be encoded as a JSON string value with proper quotation marks.
func (o *Object) Times(key string, value []time.Time, format string) *Object {
	return o.Key(key).TimesValue(value, format)
}

// TimesValue appends an array of time.Time values to the current key in the JSON object.
// All times are formatted as strings according to the specified format.
//
// Example:
//
//	obj.Key("timestamps").TimesValue([]time.Time{time.Now(), time.Now().Add(-24*time.Hour)}, time.RFC3339)
//
// Each time will be encoded as a JSON string value with proper quotation marks.
func (o *Object) TimesValue(value []time.Time, format string) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, value time.Time) []byte {
		return appendTime(buf, value, format)
	})
	o.buf = append(o.buf, ',')
	return o
}

// Duration appends a time.Duration key-value pair to the JSON object.
//
// IMPORTANT: Unlike other numeric types, durations are encoded as strings using
// the Duration.String() representation (e.g., "1h2m3s"), not as numeric nanoseconds.
// This provides better human readability but may require specific parsing on the receiving end.
//
// If you want to represent a duration as a numeric value you can use the Int64(key, value) function
// as a time.Duration is just an alias for int64.
//
// Example:
//
//	obj.Duration("timeout", 5*time.Minute) // Encodes as "timeout":"5m0s"
func (o *Object) Duration(key string, value time.Duration) *Object {
	return o.Key(key).DurationValue(value)
}

// DurationValue appends a time.Duration value to the current key in the JSON object.
//
// IMPORTANT: Unlike other numeric types, durations are encoded as strings using
// the Duration.String() representation (e.g., "1h2m3s"), not as numeric nanoseconds.
// This provides better human readability but may require specific parsing on the receiving end.
//
// If you want to represent a duration as a numeric value you can use the Int64Value(value) function
// as a time.Duration is just an alias for int64.
//
// Example:
//
//	obj.Key("timeout").DurationValue(5*time.Minute) // Encodes as "timeout":"5m0s"
func (o *Object) DurationValue(value time.Duration) *Object {
	return o.StringValue(value.String())
}

// Durations appends an array of time.Duration values as a key-value pair to the JSON object.
//
// IMPORTANT: Unlike other numeric types, durations are encoded as strings using
// the Duration.String() representation (e.g., "1h2m3s"), not as numeric nanoseconds.
// This provides better human readability but may require specific parsing on the receiving end.
//
// Example:
//
//	obj.Durations("intervals", []time.Duration{5*time.Second, 10*time.Minute})
//	// Encodes as "intervals":["5s","10m0s"]
func (o *Object) Durations(key string, value []time.Duration) *Object {
	return o.Key(key).DurationsValue(value)
}

// DurationsValue appends an array of time.Duration values to the current key in the JSON object.
//
// IMPORTANT: Unlike other numeric types, durations are encoded as strings using
// the Duration.String() representation (e.g., "1h2m3s"), not as numeric nanoseconds.
// This provides better human readability but may require specific parsing on the receiving end.
//
// Example:
//
//	obj.Key("intervals").DurationsValue([]time.Duration{5*time.Second, 10*time.Minute})
//	// Encodes as "intervals":["5s","10m0s"]
func (o *Object) DurationsValue(value []time.Duration) *Object {
	o.buf = appendArray(o.buf, value, func(buf []byte, v time.Duration) []byte {
		return appendString(buf, v.String())
	})
	o.buf = append(o.buf, ',')
	return o
}

// Object adds a new nested object with the given key.
// This is a convenience method that combines Key() and StartObject().
//
// Example:
//
//	obj.Object("person").
//	    String("name", "John").
//	    Int("age", 30).
//	EndObject()
//
// Don't forget to call EndObject() when you're done adding properties to the nested object.
func (o *Object) Object(key string) *Object {
	return o.Key(key).StartObject()
}

// StartObject begins a new JSON object without a key.
// This is typically used after Array() or StartArray() when adding objects to an array.
//
// Example:
//
//	obj.Array("people").
//	    StartObject().
//	        String("name", "Alice").
//	    EndObject().
//	    StartObject().
//	        String("name", "Bob").
//	    EndObject().
//	EndArray()
//
// Don't forget to call EndObject() when you're done adding properties to the object.
func (o *Object) StartObject() *Object {
	o.buf = append(o.buf, '{')
	return o
}

// EndObject completes the current object by adding a closing brace.
// It should be called after Object()/StartObject() to close the nested object.
//
// If the object is empty it will output "{}".
//
// IMPORTANT: Each call to Object()/StartObject() must be paired with a call to EndObject().
// Unbalanced calls may result in invalid JSON.
func (o *Object) EndObject() *Object {
	// If the object is empty just append the closing tag
	// else replace the final comma with the closing tag
	if o.buf[len(o.buf)-1] == '{' {
		o.buf = append(o.buf, '}')
	} else {
		o.buf[len(o.buf)-1] = '}'
	}

	o.buf = append(o.buf, ',')
	return o
}

// Array adds a new array with the given key.
// This is a convenience method that combines Key() and StartArray().
//
// Example:
//
//	obj.Array("numbers").
//	    StartObject().Int("value", 1).EndObject().
//	    StartObject().Int("value", 2).EndObject().
//	EndArray()
//
// The difference between this and methods like Ints() is that Array allows
// you to create arrays of heterogeneous or complex objects, while type-specific
// methods like Ints() are for arrays where all elements are of the same type.
//
// Don't forget to call EndArray() when you're done adding items to the array.
func (o *Object) Array(key string) *Object {
	return o.Key(key).StartArray()
}

// StartArray begins a new JSON array without a key.
// This is typically used for creating nested array structures.
//
// Example:
//
//	obj.Array("matrix").
//	    StartArray().Int(1).Int(2).EndArray().
//	    StartArray().Int(3).Int(4).EndArray().
//	EndArray()
//
// Don't forget to call EndArray() when you're done adding items to the array.
func (o *Object) StartArray() *Object {
	o.buf = append(o.buf, '[')
	return o
}

// EndArray completes the current array by adding a closing bracket.
// It should be called after Array()/StartArray() to close the array.
//
// If the array is empty (no items added), it will output "[]".
// Otherwise, it will replace the trailing comma with a closing bracket.
//
// IMPORTANT: Each call to Array()/StartArray() must be paired with a call to EndArray().
// Unbalanced calls may result in invalid JSON.
func (o *Object) EndArray() *Object {
	// If the array is empty just append the closing array tag
	// otherwise replace the final , with a closing array tag
	if o.buf[len(o.buf)-1] == '[' {
		o.buf = append(o.buf, ']')
	} else {
		o.buf[len(o.buf)-1] = ']'
	}

	o.buf = append(o.buf, ',')
	return o
}

// Build finalizes the JSON object and returns the resulting byte slice.
// This should be called once, after all key-value pairs have been added.
//
// If the object is empty, it returns "{}".
//
// Example:
//
//	buf := make([]byte, 0, 1024)
//	json := fson.NewObject(buf).
//	    String("name", "John").
//	    Int("age", 30).
//	    Build()
//	fmt.Println(string(json))
//
// IMPORTANT: The returned byte slice references the same underlying memory as
// the input buffer. If you need to reuse the buffer for another JSON object,
// make sure to copy the result first or process it before reusing the buffer.
func (o *Object) Build() []byte {
	if o.buf[len(o.buf)-1] != '{' {
		o.buf[len(o.buf)-1] = '}'
		return o.buf
	}

	o.buf = append(o.buf, '}')
	return o.buf
}

func appendString(buf []byte, s string) []byte {
	buf = append(buf, '"')
	buf = safeAppendString(
		utf8.DecodeRuneInString,
		buf,
		s,
	)
	return append(buf, '"')
}

// The hex characters.
const _hex = "0123456789abcdef"

// safeAppendString is a generic "append to buffer" implementation that handles string escaping for JSON encoding.
//
// This function processes a string-like value (either []byte or string) and properly escapes
// all special characters according to JSON syntax rules. It efficiently handles:
//
// - UTF-8 encoded text with possible invalid sequences
// - JSON escape sequences for quotes, backslashes, and control characters
// - Unicode characters beyond the ASCII range
// - Special JSON escape sequences (\", \\, \n, \r, \t)
// - Control characters (ASCII < 0x20)
//
// This function is heavily inspired by the zapcore library used by uber's Zap logging framework.
//
// The function implementation can be summarized as follows.
//  1. It scans through the input, skipping characters that don't need escaping
//  2. When it finds a character requiring special handling, it copies all previously
//     accumulated "safe" characters in a single operation, then handles the special case
//  3. This approach avoids allocating intermediary strings and minimizes copy operations
//
// Parameters:
//   - appendTo: Function to append string-like content to the byte buffer
//   - decodeRune: Function to decode the next rune in the string-like content
//   - buf: Destination buffer where the escaped string will be appended
//   - s: Source string-like content to be escaped
//
// Returns:
//   - The updated buffer with the escaped string appended
func safeAppendString[S []byte | string](decodeRune func(S) (rune, int), buf []byte, s S) []byte { //nolint: cyclop
	lastProcessedIndex := 0

	// Process the entire string
	for currentIndex := 0; currentIndex < len(s); {
		// Handle multibyte UTF-8 characters
		if s[currentIndex] >= utf8.RuneSelf {
			// Decode the rune to handle it properly
			r, runeSize := decodeRune(s[currentIndex:])

			// Found an invalid UTF-8 sequence, handle it by replacing it with
			// the UTF8 replacement character (utf8.RuneError)
			if r == utf8.RuneError && runeSize == 1 {
				buf = append(buf, s[lastProcessedIndex:currentIndex]...)
				buf = utf8.AppendRune(buf, utf8.RuneError)

				currentIndex++
				lastProcessedIndex = currentIndex
				continue
			}

			// Happy path just continue
			currentIndex += runeSize
			continue
		}

		// Handle ASCII characters (smaller than 128)
		// Character doesn't need escaping increment index and continue
		if s[currentIndex] >= 0x20 && s[currentIndex] != '\\' && s[currentIndex] != '"' {
			currentIndex++
			continue
		}

		// Character needs escaping - handle accumulated safe characters first
		buf = append(buf, s[lastProcessedIndex:currentIndex]...)

		// Apply appropriate escaping based on character type
		switch s[currentIndex] {
		case '\\', '"':
			// Backslash and quote need a backslash prefix
			buf = append(buf, '\\', s[currentIndex])
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		default:
			// Control characters (ASCII < 0x20) use \u00XX format
			buf = append(buf, `\u00`...)
			buf = append(buf, _hex[s[currentIndex]>>4])
			buf = append(buf, _hex[s[currentIndex]&0xF])
		}

		// As always, increment and continue
		currentIndex++
		lastProcessedIndex = currentIndex
	}

	// Append any remaining unprocessed characters
	return append(buf, s[lastProcessedIndex:]...)
}

func appendTime(buf []byte, t time.Time, format string) []byte {
	buf = append(buf, '"')
	buf = t.AppendFormat(buf, format)
	return append(buf, '"')
}

// appendFloat appends the provided float to the provided buffer.
func appendFloat(buff []byte, val float64, bitSize int) []byte {
	switch {
	case math.IsNaN(val):
		return appendString(buff, "NaN")
	case math.IsInf(val, 1):
		return appendString(buff, "+Inf")
	case math.IsInf(val, -1):
		return appendString(buff, "-Inf")
	default:
		return strconv.AppendFloat(buff, val, 'f', -1, bitSize)
	}
}

// appendArray appends an array of provided elements of type T.
func appendArray[T any](buf []byte, vals []T, appendFn func([]byte, T) []byte) []byte {
	// If the array is empty, return the empty array marker
	if len(vals) == 0 {
		return append(buf, '[', ']')
	}

	// Open the array brackets
	buf = append(buf, '[')

	// Append the first element
	buf = appendFn(buf, vals[0])

	// Append the rest of the elements
	for _, val := range vals[1:] {
		buf = appendFn(append(buf, ','), val)
	}

	// Close the array brackets
	return append(buf, ']')
}
