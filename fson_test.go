package fson_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/LucasRouckhout/fson"
	"github.com/LucasRouckhout/fson/fsonutil"
	"math"
	"testing"
	"time"
	"unicode/utf8"
)

var buffPool = fsonutil.NewPool()

func FuzzJsonObject(f *testing.F) {
	f.Add("data", true, 42, int8(8), int16(16), int32(32), int64(64),
		uint(42), uint8(8), uint16(16), uint32(32), uint64(64),
		float32(3.14), float64(2.71), int64(1620000000), int64(5000000000))

	f.Fuzz(func(t *testing.T, str string, bl bool, i int, i8 int8, i16 int16, i32 int32, i64 int64, ui uint, ui8 uint8, ui16 uint16, ui32 uint32, ui64 uint64, f32 float32, f64 float64, timeUnix int64, durationNano int64) {
		// Convert the int64 values to time.Time and time.Duration
		tm := time.Unix(timeUnix, 0)
		dur := time.Duration(durationNano)

		// Get a buffer from the pool
		buf := buffPool.Get()
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
		b := fson.NewObject(buf.Bytes()).
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
	buf := buffPool.Get()
	defer buffPool.Put(buf)

	b := fson.NewObject(buf.Bytes()).String("foo", "bar").Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}
	if !utf8.Valid(b) {
		t.Errorf("invalid utf8: %s", b)
	}
}

func TestObject_Null(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get()
	defer buffPool.Put(buf)

	b := fson.NewObject(buf.Bytes()).Null("foo").Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}
	if !utf8.Valid(b) {
		t.Errorf("invalid utf8: %s", b)
	}
}

func TestObject_NullArray(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get()
	defer buffPool.Put(buf)

	obj := fson.NewObject(buf.Bytes())
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
	buf := buffPool.Get()
	defer buffPool.Put(buf)

	// Test populated array
	strings := []string{"hello", "world", "!"}
	b := fson.NewObject(buf.Bytes()).Strings("foo", strings).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}
	if !utf8.Valid(b) {
		t.Errorf("invalid utf8: %s", b)
	}
}

func TestObject_StringsEmpty(t *testing.T) {
	t.Parallel()
	buf := buffPool.Get()
	defer buffPool.Put(buf)

	var emptyStrings []string
	b := fson.NewObject(buf.Bytes()).Strings("foo", emptyStrings).Build()

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

	buf := buffPool.Get()
	defer buffPool.Put(buf)

	b := fson.NewObject(buf.Bytes()).Bool("foo", true).Bool("bar", false).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (bool): %s", b)
	}
}

// Test for Bools array
func TestObject_Bools(t *testing.T) {
	t.Parallel()

	buf := buffPool.Get()
	defer buffPool.Put(buf)

	// Test populated array
	bools := []bool{true, false, true}
	b := fson.NewObject(buf.Bytes()).Bools("foo", bools).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}
}

func TestObject_BoolsEmpty(t *testing.T) {
	t.Parallel()

	buf := buffPool.Get()
	defer buffPool.Put(buf)

	var emptyBools []bool
	b := fson.NewObject(buf.Bytes()).Bools("foo", emptyBools).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json (empty array): %s", b)
	}
}

// Test for Int
func TestObject_Int(t *testing.T) {
	t.Parallel()

	buf := buffPool.Get()
	defer buffPool.Put(buf)

	b := fson.NewObject(buf.Bytes()).
		Int("foo", 42).
		Int("bar", 0).
		Int("bar", -42).
		Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
	}
}

// Test for Ints array
func TestObject_Ints(t *testing.T) {
	t.Parallel()

	buf := buffPool.Get()
	defer buffPool.Put(buf)

	// Test populated array
	ints := []int{1, 0, -1, 42}
	var emptyInts []int
	b := fson.NewObject(buf.Bytes()).Ints("foo", ints).Ints("bar", emptyInts).Build()

	if !json.Valid(b) {
		t.Errorf("invalid json: %s", b)
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
	buf := buffPool.Get()
	defer buffPool.Put(buf)

	// Test completely empty object
	b := fson.NewObject(buf.Bytes()).Build()

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

	buf := buffPool.Get()
	defer buffPool.Put(buf)

	// Test strings with special characters that should be escaped
	b := fson.NewObject(buf.Bytes()).
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
}

// Test for key name escaping
func TestObject_KeyNameEscaping(t *testing.T) {
	t.Parallel()

	buf := buffPool.Get()
	defer buffPool.Put(buf)

	// Test keys with characters that should be escaped
	b := fson.NewObject(buf.Bytes()).
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

	buf := buffPool.Get()
	defer buffPool.Put(buf)

	// First use
	b1 := fson.NewObject(buf.Bytes()).String("key1", "value1").Build()
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
}

func TestObject_Reset(t *testing.T) {
	t.Parallel()

	buf := buffPool.Get()
	defer buffPool.Put(buf)

	obj := fson.NewObject(buf.Bytes())

	b1 := obj.String("foo", "bar").Build()
	json1 := string(b1)

	if json1 != `{"foo":"bar"}` {
		t.Errorf("first JSON incorrect: %s", json1)
	}

	// Reset buffer
	obj.Reset()

	b2 := obj.String("bar", "foo").Build()
	json2 := string(b2)

	if json2 != `{"bar":"foo"}` {
		t.Errorf("second JSON incorrect: %s", json2)
	}
}

var result []byte

func BenchmarkObject_BuildSimple(b *testing.B) {
	buf := make([]byte, 1024*100)

	var r []byte
	obj := fson.NewObject(buf)
	for b.Loop() {
		r = obj.String("foo", "bar").Build()
		obj.Reset()
	}

	result = r
}

func BenchmarkJson_StdlibSimple(b *testing.B) {
	type A struct {
		Foo string `json:"foo"`
	}

	a := A{Foo: "bar"}
	var r []byte

	buf := make([]byte, 1024*100)
	buffer := bytes.NewBuffer(buf)

	for b.Loop() {
		_ = json.NewEncoder(buffer).Encode(&a)
		r = buffer.Bytes()
		buffer.Reset()
	}

	result = r
}

func BenchmarkObject_BuildComplex(b *testing.B) {
	buf := make([]byte, 1024*100)

	var r []byte
	obj := fson.NewObject(buf)
	for b.Loop() {
		r = obj.
			String("unicode", "😀😀😀😀😀😀😀😀😀😀😀aqwdqwd😀😀😀😀").
			Object("obj").
			Float64("float", 1.123124313).
			Array("items").
			StringValue("first").
			NullValue(). // Add a null value in the array
			StringValue("third").
			EndArray().
			Build()
		obj.Reset()
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

	buf := make([]byte, 1024*100)
	buffer := bytes.NewBuffer(buf)

	for b.Loop() {
		_ = json.NewEncoder(buffer).Encode(&complexStruct)
		r = buffer.Bytes()
		buffer.Reset()
	}

	result = r
}

func BenchmarkObject_BuildLarge(b *testing.B) {
	buf := make([]byte, 1024*100)

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
	obj := fson.NewObject(buf)
	for b.Loop() {
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
		obj.Reset()
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

	buf := make([]byte, 1024*100)
	buffer := bytes.NewBuffer(buf)

	for b.Loop() {
		_ = json.NewEncoder(buffer).Encode(&largeStruct)
		r = buffer.Bytes()
		buffer.Reset()
	}

	result = r
}
