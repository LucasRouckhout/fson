package fson_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/LucasRouckhout/fson"
	"math"
	"sync"
	"testing"
	"time"
	"unicode/utf8"
)

var buffPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 100)
	},
}

func FuzzJsonObject(f *testing.F) {
	f.Add("data", true, 42, int8(8), int16(16), int32(32), int64(64),
		uint(42), uint8(8), uint16(16), uint32(32), uint64(64),
		float32(3.14), float64(2.71), int64(1620000000), int64(5000000000))

	f.Fuzz(func(t *testing.T, str string, bl bool, i int, i8 int8, i16 int16, i32 int32, i64 int64, ui uint, ui8 uint8, ui16 uint16, ui32 uint32, ui64 uint64, f32 float32, f64 float64, timeUnix int64, durationNano int64) {
		// Convert the int64 values to time.Time and time.Duration
		tm := time.Unix(timeUnix, 0)
		dur := time.Duration(durationNano)

		// Get a buffer from the pool
		buf := buffPool.Get().([]byte)
		defer buffPool.Put(buf)

		// Create populated arrays of each type
		strings := []string{str, "test", ""}
		bools := []bool{bl, true, false}
		ints := []int{i, 0, -1, 100}
		ints8 := []int8{i8, 0, -1, 100}
		ints16 := []int16{i16, 0, -1, 100}
		ints32 := []int32{i32, 0, -1, 100}
		ints64 := []int64{i64, 0, -1, 100}
		uints := []uint{ui, 0, 1, 100}
		uints8 := []uint8{ui8, 0, 1, 100}
		uints16 := []uint16{ui16, 0, 1, 100}
		uints32 := []uint32{ui32, 0, 1, 100}
		uints64 := []uint64{ui64, 0, 1, 100}
		floats32 := []float32{f32, 0, -1.5, 3.14}
		floats64 := []float64{f64, 0, -1.5, 3.14}
		times := []time.Time{tm, time.Unix(0, 0), time.Now()}
		durations := []time.Duration{dur, time.Second, time.Hour}

		// Create empty arrays
		emptyStrings := []string{}
		emptyBools := []bool{}
		emptyInts := []int{}
		emptyInts8 := []int8{}
		emptyInts16 := []int16{}
		emptyInts32 := []int32{}
		emptyInts64 := []int64{}
		emptyUints := []uint{}
		emptyUints8 := []uint8{}
		emptyUints16 := []uint16{}
		emptyUints32 := []uint32{}
		emptyUints64 := []uint64{}
		emptyFloats32 := []float32{}
		emptyFloats64 := []float64{}
		emptyTimes := []time.Time{}
		emptyDurations := []time.Duration{}

		// Build the JSON object
		b := fson.NewObject(buf).
			// Single values
			String("string", str).
			Bool("bool", bl).
			Int("int", i).
			Int8("int8", i8).
			Int16("int16", i16).
			Int32("int32", i32).
			Int64("int64", i64).
			Uint("uint", ui).
			Uint8("uint8", ui8).
			Uint16("uint16", ui16).
			Uint32("uint32", ui32).
			Uint64("uint64", ui64).
			Float32("float32", f32).
			Float64("float64", f64).
			Time("time", tm, time.RFC3339).
			Duration("duration", dur).

			// Populated arrays
			Strings("strings", strings).
			Bools("bools", bools).
			Ints("ints", ints).
			Ints8("ints8", ints8).
			Ints16("ints16", ints16).
			Ints32("ints32", ints32).
			Ints64("ints64", ints64).
			Uints("uints", uints).
			Uints8("uints8", uints8).
			Uints16("uints16", uints16).
			Uints32("uints32", uints32).
			Uints64("uints64", uints64).
			Floats32("floats32", floats32).
			Floats64("floats64", floats64).
			Times("times", times, time.RFC3339).
			Durations("durations", durations).

			// Empty arrays
			Strings("emptyStrings", emptyStrings).
			Bools("emptyBools", emptyBools).
			Ints("emptyInts", emptyInts).
			Ints8("emptyInts8", emptyInts8).
			Ints16("emptyInts16", emptyInts16).
			Ints32("emptyInts32", emptyInts32).
			Ints64("emptyInts64", emptyInts64).
			Uints("emptyUints", emptyUints).
			Uints8("emptyUints8", emptyUints8).
			Uints16("emptyUints16", emptyUints16).
			Uints32("emptyUints32", emptyUints32).
			Uints64("emptyUints64", emptyUints64).
			Floats32("emptyFloats32", emptyFloats32).
			Floats64("emptyFloats64", emptyFloats64).
			Times("emptyTimes", emptyTimes, time.RFC3339).
			Durations("emptyDurations", emptyDurations).

			// Nested object with both single values and arrays
			Object("nestedObject").
			String("string", str).
			Bool("bool", bl).
			Strings("strings", strings).
			Bools("bools", bools).
			Ints("ints", ints).
			Floats64("floats64", floats64).
			Times("times", times, time.RFC3339).

			// Test nested empty arrays
			Strings("emptyStrings", emptyStrings).
			Ints("emptyInts", emptyInts).

			// Double nested object
			Object("doubleNested").
			String("string", str).
			Ints("ints", ints).
			Strings("emptyStrings", emptyStrings).
			EndObject().
			EndObject().
			Build()

		// Check if we produced valid JSON
		if !json.Valid(b) {
			t.Errorf("invalid json: %s", b)
		}

		// Check if the output is valid UTF-8
		if !utf8.Valid(b) {
			t.Errorf("invalid utf8: %s", b)
		}
	})
}

// Test for String
func TestObject_String(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).String("foo", "bar").Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}
	if !utf8.Valid(b) {
		t.Errorf("invalid utf8: %s", b)
	}
}

func TestObject_Null(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Null("foo").Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}
	if !utf8.Valid(b) {
		t.Errorf("invalid utf8: %s", b)
	}
}

func TestObject_NullArray(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	obj := fson.NewObject(buf)
	b := obj.Array("items").
		StringValue("first").
		NullValue(). // Add a null value in the array
		StringValue("third").
		EndArray().
		Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}
	if !utf8.Valid(b) {
		t.Errorf("invalid utf8: %s", b)
	}
}

// Test for Strings array
func TestObject_Strings(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	strings := []string{"hello", "world", "!"}
	b := fson.NewObject(buf).Strings("foo", strings).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}
	if !utf8.Valid(b) {
		t.Errorf("invalid utf8: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyStrings := []string{}
	b = fson.NewObject(buf).Strings("foo", emptyStrings).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
	if !utf8.Valid(b) {
		t.Errorf("invalid utf8 (empty array): %s", b)
	}
}

// Test for Bool
func TestObject_Bool(t *testing.T) {
	t.Parallel()

	// Test true value
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)
	b := fson.NewObject(buf).Bool("foo", true).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (true): %s", b)
	}

	// Test false value
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Bool("foo", false).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (false): %s", b)
	}
}

// Test for Bools array
func TestObject_Bools(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	bools := []bool{true, false, true}
	b := fson.NewObject(buf).Bools("foo", bools).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyBools := []bool{}
	b = fson.NewObject(buf).Bools("foo", emptyBools).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Int
func TestObject_Int(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Int("foo", 42).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Int("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}

	// Test negative
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Int("foo", -42).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (negative): %s", b)
	}
}

// Test for Ints array
func TestObject_Ints(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	ints := []int{1, 0, -1, 42}
	b := fson.NewObject(buf).Ints("foo", ints).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyInts := []int{}
	b = fson.NewObject(buf).Ints("foo", emptyInts).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Int8
func TestObject_Int8(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Int8("foo", 8).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Int8("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}

	// Test negative
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Int8("foo", -8).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (negative): %s", b)
	}
}

// Test for Ints8 array
func TestObject_Ints8(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	ints8 := []int8{1, 0, -1, 8}
	b := fson.NewObject(buf).Ints8("foo", ints8).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyInts8 := []int8{}
	b = fson.NewObject(buf).Ints8("foo", emptyInts8).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Int16
func TestObject_Int16(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Int16("foo", 16).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Int16("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}

	// Test negative
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Int16("foo", -16).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (negative): %s", b)
	}
}

// Test for Ints16 array
func TestObject_Ints16(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	ints16 := []int16{1, 0, -1, 16}
	b := fson.NewObject(buf).Ints16("foo", ints16).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyInts16 := []int16{}
	b = fson.NewObject(buf).Ints16("foo", emptyInts16).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Int32
func TestObject_Int32(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Int32("foo", 32).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Int32("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}

	// Test negative
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Int32("foo", -32).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (negative): %s", b)
	}
}

// Test for Ints32 array
func TestObject_Ints32(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	ints32 := []int32{1, 0, -1, 32}
	b := fson.NewObject(buf).Ints32("foo", ints32).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyInts32 := []int32{}
	b = fson.NewObject(buf).Ints32("foo", emptyInts32).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Int64
func TestObject_Int64(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Int64("foo", 64).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Int64("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}

	// Test negative
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Int64("foo", -64).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (negative): %s", b)
	}

	// Test large number
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Int64("foo", 9223372036854775807).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (large number): %s", b)
	}
}

// Test for Ints64 array
func TestObject_Ints64(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	ints64 := []int64{1, 0, -1, 64, 9223372036854775807}
	b := fson.NewObject(buf).Ints64("foo", ints64).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyInts64 := []int64{}
	b = fson.NewObject(buf).Ints64("foo", emptyInts64).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Uint
func TestObject_Uint(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Uint("foo", 42).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Uint("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}
}

// Test for Uints array
func TestObject_Uints(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	uints := []uint{1, 0, 42, 100}
	b := fson.NewObject(buf).Uints("foo", uints).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyUints := []uint{}
	b = fson.NewObject(buf).Uints("foo", emptyUints).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Uint8
func TestObject_Uint8(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Uint8("foo", 8).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Uint8("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}
}

// Test for Uints8 array
func TestObject_Uints8(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	uints8 := []uint8{1, 0, 8, 255}
	b := fson.NewObject(buf).Uints8("foo", uints8).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyUints8 := []uint8{}
	b = fson.NewObject(buf).Uints8("foo", emptyUints8).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Uint16
func TestObject_Uint16(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Uint16("foo", 16).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Uint16("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}
}

// Test for Uints16 array
func TestObject_Uints16(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	uints16 := []uint16{1, 0, 16, 65535}
	b := fson.NewObject(buf).Uints16("foo", uints16).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyUints16 := []uint16{}
	b = fson.NewObject(buf).Uints16("foo", emptyUints16).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Uint32
func TestObject_Uint32(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Uint32("foo", 32).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Uint32("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}
}

// Test for Uints32 array
func TestObject_Uints32(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	uints32 := []uint32{1, 0, 32, 4294967295}
	b := fson.NewObject(buf).Uints32("foo", uints32).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyUints32 := []uint32{}
	b = fson.NewObject(buf).Uints32("foo", emptyUints32).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Uint64
func TestObject_Uint64(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Uint64("foo", 64).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Uint64("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}

	// Test large number
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Uint64("foo", 18446744073709551615).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (large number): %s", b)
	}
}

// Test for Uints64 array
func TestObject_Uints64(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	uints64 := []uint64{1, 0, 64, 18446744073709551615}
	b := fson.NewObject(buf).Uints64("foo", uints64).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyUints64 := []uint64{}
	b = fson.NewObject(buf).Uints64("foo", emptyUints64).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Float32
func TestObject_Float32(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Float32("foo", 3.14).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Float32("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}

	// Test negative
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Float32("foo", -3.14).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (negative): %s", b)
	}

	// Test special values
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Float32("nanValue", float32(math.NaN())).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (NaN): %s", b)
	}

	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Float32("posInf", float32(math.Inf(1))).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (+Inf): %s", b)
	}

	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Float32("negInf", float32(math.Inf(-1))).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (-Inf): %s", b)
	}

	f := fson.NewObject(buf).Strings("fooArr", []string{"bar", "foo"}).Build()
	if !json.Valid(f) {
		t.Errorf("invalid json: %s", f)
	}
}

// Test for Floats32 array
func TestObject_Floats32(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	floats32 := []float32{3.14, 0, -3.14, 1.23456}
	b := fson.NewObject(buf).Floats32("foo", floats32).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyFloats32 := []float32{}
	b = fson.NewObject(buf).Floats32("foo", emptyFloats32).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}

	// Test special values
	buf = buffPool.Get().([]byte)
	specialFloats32 := []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))}
	b = fson.NewObject(buf).Floats32("special", specialFloats32).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (special values): %s", b)
	}
}

// Test for Float64
func TestObject_Float64(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	b := fson.NewObject(buf).Float64("foo", 2.71828).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Float64("foo", 0).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero): %s", b)
	}

	// Test negative
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Float64("foo", -2.71828).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (negative): %s", b)
	}

	// Test special values
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Float64("nanValue", math.NaN()).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (NaN): %s", b)
	}

	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Float64("posInf", math.Inf(1)).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (+Inf): %s", b)
	}

	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Float64("negInf", math.Inf(-1)).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (-Inf): %s", b)
	}
}

// Test for Floats64 array
func TestObject_Floats64(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	floats64 := []float64{2.71828, 0, -2.71828, 1.2345678901234}
	b := fson.NewObject(buf).Floats64("foo", floats64).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyFloats64 := []float64{}
	b = fson.NewObject(buf).Floats64("foo", emptyFloats64).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}

	// Test special values
	buf = buffPool.Get().([]byte)
	specialFloats64 := []float64{math.NaN(), math.Inf(1), math.Inf(-1)}
	b = fson.NewObject(buf).Floats64("special", specialFloats64).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (special values): %s", b)
	}
}

// Test for Time
func TestObject_Time(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	now := time.Now()
	b := fson.NewObject(buf).Time("foo", now, time.RFC3339).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test zero time
	buf = buffPool.Get().([]byte)
	zeroTime := time.Time{}
	b = fson.NewObject(buf).Time("foo", zeroTime, time.RFC3339).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (zero time): %s", b)
	}

	// Test different formats
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Time("rfc822", now, time.RFC822).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (RFC822): %s", b)
	}

	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Time("unix", now, time.UnixDate).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (UnixDate): %s", b)
	}
}

// Test for Times array
func TestObject_Times(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)
	times := []time.Time{now, past, future}
	b := fson.NewObject(buf).Times("foo", times, time.RFC3339).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyTimes := []time.Time{}
	b = fson.NewObject(buf).Times("foo", emptyTimes, time.RFC3339).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}

	// Test different formats
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).Times("rfc822", times, time.RFC822).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (RFC822): %s", b)
	}
}

// Test for Durations array
func TestObject_Durations(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test populated array
	durations := []time.Duration{time.Second, time.Minute, time.Hour, 24 * time.Hour}
	b := fson.NewObject(buf).Durations("foo", durations).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}

	// Test empty array
	buf = buffPool.Get().([]byte)
	emptyDurations := []time.Duration{}
	b = fson.NewObject(buf).Durations("foo", emptyDurations).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}

	// Test mixed durations including negative
	buf = buffPool.Get().([]byte)
	mixedDurations := []time.Duration{time.Second, 0, -time.Hour, 24 * time.Hour}
	b = fson.NewObject(buf).Durations("mixed", mixedDurations).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (mixed durations): %s", b)
	}
}

// Test for nested objects
func TestObject_NestedObjects(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test simple nested object
	b := fson.NewObject(buf).
		Object("nested").
		String("foo", "bar").
		Int("num", 42).
		EndObject().
		Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (simple nested): %s", b)
	}

	// Test empty nested object
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).
		Object("empty").
		EndObject().
		Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty nested): %s", b)
	}

	// Test multiple nested objects
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).
		Object("first").
		String("name", "first").
		EndObject().
		Object("second").
		String("name", "second").
		EndObject().
		Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (multiple nested): %s", b)
	}

	// Test deeply nested objects
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).
		Object("level1").
		Object("level2").
		Object("level3").
		String("deep", "value").
		EndObject().
		EndObject().
		EndObject().
		Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (deeply nested): %s", b)
	}
}

// Test for complex object with mixed types
func TestObject_ComplexObject(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test a complex object with various types
	now := time.Now()
	strings := []string{"hello", "world"}
	ints := []int{1, 2, 3}

	b := fson.NewObject(buf).
		String("string", "value").
		Bool("bool", true).
		Int("int", 42).
		Float64("float", 3.14159).
		Time("time", now, time.RFC3339).
		Duration("duration", time.Hour).
		Strings("strings", strings).
		Ints("ints", ints).
		Object("nested").
		String("name", "nested object").
		Bool("active", true).
		Object("deeper").
		String("level", "deep").
		EndObject().
		EndObject().
		Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (complex object): %s", b)
	}

	if !utf8.Valid(b) {
		t.Errorf("invalid utf8 (complex object): %s", b)
	}
}

func TestObject_ArrayOfObjects(t *testing.T) {
	t.Parallel()

	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	type A struct {
		String string
		Int    int
	}

	arr := []A{
		{String: "hello", Int: 42},
		{String: "world", Int: 0},
		{String: "hello", Int: 0},
	}

	obj := fson.NewObject(buf)

	obj = obj.Array("objects")
	for _, a := range arr {
		obj = obj.StartObject().
			String("string", a.String).
			Int("int", a.Int).
			EndObject()
	}
	b := obj.EndArray().Build() // -> {"objects":[{"string":"hello","int":42},{"string":"world","int":0},{"string":"hello","int":0}]}

	if !json.Valid(b) {
		t.Errorf("invalid json (array of objects): %s", b)
	}

	if !utf8.Valid(b) {
		t.Errorf("invalid utf8 (array of objects): %s", b)
	}
}

func TestObject_Floats64_WithNan(t *testing.T) {
	t.Parallel()

	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test array with various special floating point values
	specialFloats := []float64{
		1.23,                        // Regular number
		0.0,                         // Zero
		-4.56,                       // Negative number
		math.NaN(),                  // NaN (Not a Number)
		math.Inf(1),                 // Positive Infinity
		math.Inf(-1),                // Negative Infinity
		math.MaxFloat64,             // Maximum representable float64
		math.SmallestNonzeroFloat64, // Smallest positive non-zero float64
	}

	// Encode the array
	obj := fson.NewObject(buf).Floats64("special", specialFloats).Build()

	// Check if the result is valid JSON
	if !json.Valid(obj) {
		t.Errorf("expected valid JSON, got invalid JSON: %s", obj)
	}
}

func TestObject_Floats64_SkippingSpecialValues(t *testing.T) {
	// Test array with various special floating point values
	t.Parallel()
	buf := make([]byte, 0, 256)

	// Create a slice with regular and special float values
	specialFloats := []float64{
		1.23,                        // Regular number
		0.0,                         // Zero
		-4.56,                       // Negative number
		math.NaN(),                  // NaN (Not a Number) - should be skipped
		math.Inf(1),                 // Positive Infinity - should be skipped
		math.Inf(-1),                // Negative Infinity - should be skipped
		math.MaxFloat64,             // Maximum representable float64
		math.SmallestNonzeroFloat64, // Smallest positive non-zero float64
	}

	// Create JSON using the raw approach that skips special values
	obj := fson.NewObject(buf)
	obj.Key("filtered").StartArray()
	for _, v := range specialFloats {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			// Skip NaN and Infinity values
			continue
		}
		obj.Float64Value(v)
	}
	obj.EndArray()

	// For comparison, also create a regular array with all values
	obj.Floats64("all", specialFloats)

	// Build the final JSON
	result := obj.Build()

	// Verify the result is valid JSON
	if !json.Valid(result) {
		t.Errorf("expected valid JSON, got invalid JSON: %s", result)
	}

	// Unmarshal and check the filtered array
	var parsed map[string]interface{}
	err := json.Unmarshal(result, &parsed)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Check filtered array length (should have 5 elements, not 8)
	filtered, ok := parsed["filtered"].([]interface{})
	if !ok {
		t.Fatalf("expected 'filtered' to be an array")
	}

	if len(filtered) != 5 {
		t.Errorf("expected filtered array to have 5 elements (special values skipped), got %d", len(filtered))
	}

	// Check that all elements in filtered are numbers (no strings)
	for i, val := range filtered {
		if _, ok := val.(float64); !ok {
			t.Errorf("expected element %d in filtered array to be a number, got %T", i, val)
		}
	}

	// Check that the regular array has all 8 elements with mixed types
	all, ok := parsed["all"].([]interface{})
	if !ok {
		t.Fatalf("expected 'all' to be an array")
	}

	if len(all) != 8 {
		t.Errorf("expected complete array to have 8 elements, got %d", len(all))
	}

	// The regular array should have some string elements (for NaN, +Inf, -Inf)
	hasStrings := false
	for _, val := range all {
		if _, ok := val.(string); ok {
			hasStrings = true
			break
		}
	}

	if !hasStrings {
		t.Errorf("expected complete array to have string elements for special values")
	}

	// Log the result for inspection
	t.Logf("Filtered JSON array: %s", result)
}

// Test for empty object
func TestObject_EmptyObject(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test completely empty object
	b := fson.NewObject(buf).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty object): %s", b)
	}

	// Check that it's actually "{}"
	if string(b) != "{}" {
		t.Errorf("expected empty object to be {}, got: %s", b)
	}
}

// Test for special string characters
func TestObject_SpecialStringCharacters(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test strings with special characters that should be escaped
	b := fson.NewObject(buf).
		String("quotes", "with \"quotes\"").
		String("backslash", "with \\backslash").
		String("newline", "with \nnewline").
		String("tab", "with \ttab").
		String("unicode", "with unicode 😀").
		String("control", "with control \u0001 char").
		Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (special characters): %s", b)
	}

	if !utf8.Valid(b) {
		t.Errorf("invalid utf8 (special characters): %s", b)
	}

	// Test with empty string
	buf = buffPool.Get().([]byte)
	b = fson.NewObject(buf).
		String("empty", "").
		Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty string): %s", b)
	}
}

// Test for key name escaping
func TestObject_KeyNameEscaping(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Test keys with characters that should be escaped
	b := fson.NewObject(buf).
		String("with \"quotes\"", "value").
		String("with \\backslash", "value").
		String("with \nnewline", "value").
		String("with \ttab", "value").
		String("with unicode 😀", "value").
		Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (escaped keys): %s", b)
	}

	if !utf8.Valid(b) {
		t.Errorf("invalid utf8 (escaped keys): %s", b)
	}
}

// Test for edge cases with buffer reuse
func TestObject_BufferReuse(t *testing.T) {
	t.Parallel()

	// Create a buffer and use it multiple times
	buf := buffPool.Get().([]byte)

	// First use
	b1 := fson.NewObject(buf).String("key1", "value1").Build()
	json1 := string(b1)

	// Second use (without putting back to the pool)
	b2 := fson.NewObject(b1).String("key2", "value2").Build()
	json2 := string(b2)

	// Check both JSONs
	if json1 != `{"key1":"value1"}` {
		t.Errorf("first JSON incorrect: %s", json1)
	}

	if json2 != `{"key2":"value2"}` {
		t.Errorf("second JSON incorrect: %s", json2)
	}

	// Clean up
	buffPool.Put(b2)
}

var result []byte

func BenchmarkObject_BuildSimple(b *testing.B) {
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	var r []byte
	for b.Loop() {
		r = fson.NewObject(buf).String("foo", "bar").Build()
	}

	result = r
}

func BenchmarkJson_StdlibSimple(b *testing.B) {
	type A struct {
		Foo string `json:"foo"`
	}

	a := A{Foo: "bar"}
	var r []byte
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	buffer := bytes.NewBuffer(buf)
	for b.Loop() {
		_ = json.NewEncoder(buffer).Encode(&a)
		r = buffer.Bytes()
	}

	result = r
}

func BenchmarkObject_BuildComplex(b *testing.B) {
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	var r []byte
	for b.Loop() {
		r = fson.NewObject(buf).
			String("unicode", "😀😀😀😀😀😀😀😀😀😀😀aqwdqwd😀😀😀😀").
			Object("obj").
			Float64("float", 1.123124313).
			Array("items").
			StringValue("first").
			NullValue(). // Add a null value in the array
			StringValue("third").
			EndArray().
			Build()
	}
	result = r
}

func BenchmarkJson_StdlibComplex(b *testing.B) {
	type Item struct {
		Items []interface{} `json:"items"`
		Float float64       `json:"float"`
	}

	type ComplexStruct struct {
		Unicode string `json:"unicode"`
		Obj     Item   `json:"obj"`
	}

	// Create a struct with the same data as the fson example
	complexStruct := ComplexStruct{
		Unicode: "😀😀😀😀😀😀😀😀😀😀😀aqwdqwd😀😀😀😀",
		Obj: Item{
			Float: 1.123124313,
			Items: []interface{}{"first", nil, "third"},
		},
	}

	var r []byte

	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	buffer := bytes.NewBuffer(buf)
	for b.Loop() {
		_ = json.NewEncoder(buffer).Encode(&complexStruct)
		r = buffer.Bytes()
	}

	result = r
}

func BenchmarkObject_BuildLarge(b *testing.B) {
	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	// Prepare some test data
	loremIpsum := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."

	// Fix tags array creation
	tags := make([]string, 20)
	for i := 0; i < 20; i++ {
		tags[i] = fmt.Sprintf("tag-%d", i)
	}

	// Precalculate time values outside the loop
	now := time.Now()
	historyTimes := make([]time.Time, 10)
	for i := 0; i < 10; i++ {
		historyTimes[i] = now.Add(time.Duration(-i) * time.Hour)
	}

	// Precalculate formatted strings used in loops
	itemNames := make([]string, 50)
	itemActions := make([]string, 10)
	itemUsers := make([]string, 10)
	subItemLabels := make([]string, 50*5)

	for i := 0; i < 50; i++ {
		itemNames[i] = fmt.Sprintf("Item %d", i)
		for j := 0; j < 5; j++ {
			subItemLabels[i*5+j] = fmt.Sprintf("SubItem %d-%d", i, j)
		}
	}

	for i := 0; i < 10; i++ {
		itemActions[i] = fmt.Sprintf("Action %d", i)
		itemUsers[i] = fmt.Sprintf("user%d", i)
	}

	var r []byte
	for b.Loop() {
		// fson.NewObject already resets the buffer internally
		obj := fson.NewObject(buf)

		// Add a variety of scalar values
		obj.String("id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		obj.Int("count", 12345)
		obj.Float64("amount", 9876.54321)
		obj.Bool("active", true)
		obj.Time("created", now, time.RFC3339)
		obj.Null("optional")

		// Add an array of simple values
		obj.Strings("tags", tags)

		// Add a large string
		obj.String("description", loremIpsum)

		// Add an array of objects
		obj.Array("items")
		for i := 0; i < 50; i++ {
			obj.StartObject()
			obj.String("name", itemNames[i])
			obj.Int("index", i)
			obj.Float64("value", float64(i)*1.5)
			obj.Bool("selected", i%3 == 0)
			obj.Array("subItems")
			for j := 0; j < 5; j++ {
				obj.StartObject()
				obj.String("label", subItemLabels[i*5+j])
				obj.Int("priority", j)
				obj.EndObject()
			}
			obj.EndArray()
			obj.EndObject()
		}
		obj.EndArray()

		// Add a deeply nested object
		obj.Object("metadata")
		obj.String("version", "2.0.0")
		obj.Object("author")
		obj.String("name", "John Doe")
		obj.String("email", "john.doe@example.com")
		obj.Object("contact")
		obj.String("phone", "+1-555-123-4567")
		obj.Object("address")
		obj.String("street", "123 Main St")
		obj.String("city", "Anytown")
		obj.String("country", "USA")
		obj.Object("geo")
		obj.Float64("lat", 37.7749)
		obj.Float64("lng", -122.4194)
		obj.EndObject()
		obj.EndObject()
		obj.EndObject()
		obj.EndObject()
		obj.Array("history")
		for i := 0; i < 10; i++ {
			obj.StartObject()
			obj.String("action", itemActions[i])
			obj.Time("timestamp", historyTimes[i], time.RFC3339)
			obj.Object("details")
			obj.String("user", itemUsers[i])
			obj.Int("status", 200+i)
			obj.EndObject()
			obj.EndObject()
		}
		obj.EndArray()
		obj.EndObject()

		r = obj.Build()
	}

	result = r
}

func BenchmarkJson_StdlibLarge(b *testing.B) {
	// Define types matching the structure created with fson
	type GeoLocation struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}

	type Address struct {
		Street  string      `json:"street"`
		City    string      `json:"city"`
		Country string      `json:"country"`
		Geo     GeoLocation `json:"geo"`
	}

	type Contact struct {
		Phone   string  `json:"phone"`
		Address Address `json:"address"`
	}

	type Author struct {
		Name    string  `json:"name"`
		Email   string  `json:"email"`
		Contact Contact `json:"contact"`
	}

	type ActionDetail struct {
		User   string `json:"user"`
		Status int    `json:"status"`
	}

	type HistoryItem struct {
		Action    string       `json:"action"`
		Timestamp time.Time    `json:"timestamp"`
		Details   ActionDetail `json:"details"`
	}

	type Metadata struct {
		Version string        `json:"version"`
		Author  Author        `json:"author"`
		History []HistoryItem `json:"history"`
	}

	type SubItem struct {
		Label    string `json:"label"`
		Priority int    `json:"priority"`
	}

	type Item struct {
		Name     string    `json:"name"`
		Index    int       `json:"index"`
		Value    float64   `json:"value"`
		Selected bool      `json:"selected"`
		SubItems []SubItem `json:"subItems"`
	}

	type LargeStruct struct {
		ID          string    `json:"id"`
		Count       int       `json:"count"`
		Amount      float64   `json:"amount"`
		Active      bool      `json:"active"`
		Created     time.Time `json:"created"`
		Optional    any       `json:"optional"`
		Tags        []string  `json:"tags"`
		Description string    `json:"description"`
		Items       []Item    `json:"items"`
		Metadata    Metadata  `json:"metadata"`
	}

	// Create test data
	loremIpsum := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."
	now := time.Now()

	// Prepare tags
	tags := make([]string, 20)
	for i := 0; i < 20; i++ {
		tags[i] = fmt.Sprintf("tag-%d", i)
	}

	// Prepare items
	items := make([]Item, 50)
	for i := 0; i < 50; i++ {
		subItems := make([]SubItem, 5)
		for j := 0; j < 5; j++ {
			subItems[j] = SubItem{
				Label:    fmt.Sprintf("SubItem %d-%d", i, j),
				Priority: j,
			}
		}

		items[i] = Item{
			Name:     fmt.Sprintf("Item %d", i),
			Index:    i,
			Value:    float64(i) * 1.5,
			Selected: i%3 == 0,
			SubItems: subItems,
		}
	}

	// Prepare history items
	historyItems := make([]HistoryItem, 10)
	for i := 0; i < 10; i++ {
		historyItems[i] = HistoryItem{
			Action:    fmt.Sprintf("Action %d", i),
			Timestamp: now.Add(time.Duration(-i) * time.Hour),
			Details: ActionDetail{
				User:   fmt.Sprintf("user%d", i),
				Status: 200 + i,
			},
		}
	}

	// Create the large struct
	largeStruct := LargeStruct{
		ID:          "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		Count:       12345,
		Amount:      9876.54321,
		Active:      true,
		Created:     now,
		Optional:    nil,
		Tags:        tags,
		Description: loremIpsum,
		Items:       items,
		Metadata: Metadata{
			Version: "2.0.0",
			Author: Author{
				Name:  "John Doe",
				Email: "john.doe@example.com",
				Contact: Contact{
					Phone: "+1-555-123-4567",
					Address: Address{
						Street:  "123 Main St",
						City:    "Anytown",
						Country: "USA",
						Geo: GeoLocation{
							Lat: 37.7749,
							Lng: -122.4194,
						},
					},
				},
			},
			History: historyItems,
		},
	}

	var r []byte

	buf := buffPool.Get().([]byte)
	defer buffPool.Put(buf)

	buffer := bytes.NewBuffer(buf)
	for b.Loop() {
		_ = json.NewEncoder(buffer).Encode(&largeStruct)
		r = buffer.Bytes()
	}

	result = r
}
