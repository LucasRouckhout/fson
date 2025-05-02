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

// Package fson provides a simple, performant and allocation-free JSON encoder
package fson

import (
	"math"
	"strconv"
	"time"
	"unicode/utf8"
)

type Object struct {
	buf []byte
}

func NewObject(buf []byte) *Object {
	obj := &Object{
		buf[:0], // Reset buffer
	}

	obj.buf = append(obj.buf, '{')

	return obj
}

func (o *Object) String(key, value string) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendString(o.buf, value)
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Strings(key string, value []string) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value string) []byte {
		return appendString(buf, value)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Int(key string, value int) *Object {
	return o.Int64(key, int64(value))
}

func (o *Object) Ints(key string, value []int) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value int) []byte {
		return strconv.AppendInt(buf, int64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Int8(key string, value int8) *Object {
	return o.Int64(key, int64(value))
}

func (o *Object) Ints8(key string, value []int8) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value int8) []byte {
		return strconv.AppendInt(buf, int64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Int16(key string, value int16) *Object {
	return o.Int64(key, int64(value))
}

func (o *Object) Ints16(key string, value []int16) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value int16) []byte {
		return strconv.AppendInt(buf, int64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Int32(key string, value int32) *Object {
	return o.Int64(key, int64(value))
}

func (o *Object) Ints32(key string, value []int32) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value int32) []byte {
		return strconv.AppendInt(buf, int64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Int64(key string, value int64) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = strconv.AppendInt(o.buf, value, 10)
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Ints64(key string, value []int64) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value int64) []byte {
		return strconv.AppendInt(buf, value, 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Uint(key string, value uint) *Object {
	return o.Uint64(key, uint64(value))
}

func (o *Object) Uints(key string, value []uint) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value uint) []byte {
		return strconv.AppendUint(buf, uint64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Uint8(key string, value uint8) *Object {
	return o.Uint64(key, uint64(value))

}

func (o *Object) Uints8(key string, value []uint8) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value uint8) []byte {
		return strconv.AppendUint(buf, uint64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Uint16(key string, value uint16) *Object {
	return o.Uint64(key, uint64(value))

}

func (o *Object) Uints16(key string, value []uint16) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value uint16) []byte {
		return strconv.AppendUint(buf, uint64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Uint32(key string, value uint32) *Object {
	return o.Uint64(key, uint64(value))

}

func (o *Object) Uints32(key string, value []uint32) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value uint32) []byte {
		return strconv.AppendUint(buf, uint64(value), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Uint64(key string, value uint64) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = strconv.AppendUint(o.buf, value, 10)
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Uints64(key string, value []uint64) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value uint64) []byte {
		return strconv.AppendUint(buf, value, 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Float32(key string, value float32) *Object {
	return o.Float64(key, float64(value))
}

func (o *Object) Floats32(key string, value []float32) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value float32) []byte {
		return appendFloat(buf, float64(value), 32)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Float64(key string, value float64) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendFloat(o.buf, value, 64)
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Floats64(key string, value []float64) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value float64) []byte {
		return appendFloat(buf, value, 64)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Bool(key string, value bool) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = strconv.AppendBool(o.buf, value)
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Bools(key string, value []bool) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value bool) []byte {
		return strconv.AppendBool(buf, value)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Time(key string, value time.Time, format string) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendTime(o.buf, value, format)
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Times(key string, value []time.Time, format string) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, value time.Time) []byte {
		return appendTime(buf, value, format)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Duration(key string, value time.Duration) *Object {
	return o.Int64(key, value.Nanoseconds()) // TODO: Allow units to be provided
}

func (o *Object) Durations(key string, value []time.Duration) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = appendArray(o.buf, value, func(buf []byte, v time.Duration) []byte {
		return strconv.AppendInt(buf, v.Nanoseconds(), 10)
	})
	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) StartObject(key string) *Object {
	o.buf = appendKey(o.buf, key)
	o.buf = append(o.buf, '{')
	return o
}

func (o *Object) EndObject() *Object {
	// If the *Object is empty just append the closing tag
	// else replace the final comma with the closing tag
	if o.buf[len(o.buf)-1] == '{' {
		o.buf = append(o.buf, '}')
	} else {
		o.buf[len(o.buf)-1] = '}'
	}

	o.buf = append(o.buf, ',')
	return o
}

func (o *Object) Build() []byte {
	if o.buf[len(o.buf)-1] != '{' {
		o.buf[len(o.buf)-1] = '}'
		return o.buf
	}

	o.buf = append(o.buf, '}')
	return o.buf
}

func appendKey(buf []byte, key string) []byte {
	buf = appendString(buf, key)
	return append(buf, ':')
}

// appendString safely adds a byte slice to the buffer taking into account correct JSON escaping
func appendString(buf []byte, s string) []byte {
	buf = append(buf, '"')
	buf = safeAppendStringLike(
		func(buf []byte, s string) []byte {
			return append(buf, s...)
		},
		utf8.DecodeRuneInString,
		buf,
		s,
	)
	return append(buf, '"')
}

// For JSON-escaping
const _hex = "0123456789abcdef"

// safeAppendStringLike is a generic implementation of safeAddString and safeAddByteString.
// It appends a string or byte slice to the buffer, escaping all special characters.
func safeAppendStringLike[S []byte | string](
	// appendTo appends this string-like *Object to the buffer.
	appendTo func([]byte, S) []byte,
	// decodeRune decodes the next rune from the string-like *Object
	// and returns its value and width in bytes.
	decodeRune func(S) (rune, int),
	buf []byte,
	s S,
) []byte {
	// The encoding logic below works by skipping over characters
	// that can be safely copied as-is,
	// until a character is found that needs special handling.
	// At that point, we copy everything we've seen so far,
	// and then handle that special character.
	//
	// last is the index of the last byte that was copied to the buffer.
	last := 0
	for i := 0; i < len(s); {
		if s[i] >= utf8.RuneSelf {
			// Character >= RuneSelf may be part of a multibyte rune.
			// They need to be decoded before we can decide how to handle them.
			r, size := decodeRune(s[i:])
			if r != utf8.RuneError || size != 1 {
				// No special handling required.
				// Skip over this rune and continue.
				i += size
				continue
			}

			// Invalid UTF-8 sequence.
			// Replace it with the Unicode replacement character.
			appendTo(buf, s[last:i])
			buf = utf8.AppendRune(buf, utf8.RuneError)

			i++
			last = i
		} else {
			// Character < RuneSelf is a single-byte UTF-8 rune.
			if s[i] >= 0x20 && s[i] != '\\' && s[i] != '"' {
				// No escaping necessary.
				// Skip over this character and continue.
				i++
				continue
			}

			// This character needs to be escaped.
			buf = append(buf, s[last:i]...)
			switch s[i] {
			case '\\', '"':
				buf = append(buf, '\\')
				buf = append(buf, s[i])
			case '\n':
				buf = append(buf, '\\')
				buf = append(buf, 'n')
			case '\r':
				buf = append(buf, '\\')
				buf = append(buf, 'r')
			case '\t':
				buf = append(buf, '\\')
				buf = append(buf, 't')
			default:
				// Encode bytes < 0x20, except for the escape sequences above.
				buf = append(buf, `\u00`...)
				buf = append(buf, _hex[s[i]>>4])
				buf = append(buf, _hex[s[i]&0xF])
			}

			i++
			last = i
		}
	}

	// add remaining
	return appendTo(buf, s[last:])
}

func appendTime(buf []byte, t time.Time, format string) []byte {
	buf = append(buf, '"')
	buf = t.AppendFormat(buf, format)
	return append(buf, '"')
}

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

// appendArray appends an array of provided elements of type T. It requires a type specific function
// which it uses to append each individual element
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
