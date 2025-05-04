.PHONY: test
test:
	go test fson_test.go

.PHONY: fuzz
fuzz:
	go test -fuzz=FuzzJsonObject

.PHONY: fuzz_clean
fuzz_clean:
	go clean -testcache
	rm -rf testdata/fuzz

lint:
	golangci-lint run fson.go
	golangci-lint run ./fsonutil/fsonutil.go

.PHONY: benchmark

benchmark:
	@echo "Running JSON encoding benchmarks..."
	@go test -bench=. -benchmem | \
	awk 'BEGIN { \
		printf "%-30s %-15s %-15s %-15s %-15s %-15s\n", "Benchmark", "Time/Op", "Allocs/Op", "Bytes/Op", "vs Standard", "Improvement"; \
		printf "%-30s %-15s %-15s %-15s %-15s %-15s\n", "-----------------------------", "---------------", "---------------", "---------------", "---------------", "---------------"; \
		simple_fson = 0; simple_std = 0; \
		complex_fson = 0; complex_std = 0; \
		large_fson = 0; large_std = 0; \
	} \
	/BenchmarkObject_BuildSimple/ { simple_fson = $$3; simple_fson_allocs = $$5; simple_fson_bytes = $$7; } \
	/BenchmarkJson_StdlibSimple/ { \
		simple_std = $$3; simple_std_allocs = $$5; simple_std_bytes = $$7; \
		printf "%-30s %-15s %-15s %-15s %-15s %-15s\n", "BenchmarkObject_BuildSimple", simple_fson " ns/op", simple_fson_allocs " allocs/op", simple_fson_bytes " B/op", "1x (baseline)", "-"; \
		printf "%-30s %-15s %-15s %-15s %-15s %-15s\n", "BenchmarkJson_StdlibSimple", simple_std " ns/op", simple_std_allocs " allocs/op", simple_std_bytes " B/op", sprintf("%.2fx", simple_std/simple_fson), sprintf("%.2f%%", (simple_std-simple_fson)/simple_std*100); \
		printf "%-30s %-15s %-15s %-15s %-15s %-15s\n", "-----------------------------", "---------------", "---------------", "---------------", "---------------", "---------------"; \
	} \
	/BenchmarkObject_BuildComplex/ { complex_fson = $$3; complex_fson_allocs = $$5; complex_fson_bytes = $$7; } \
	/BenchmarkJson_StdlibComplex/ { \
		complex_std = $$3; complex_std_allocs = $$5; complex_std_bytes = $$7; \
		printf "%-30s %-15s %-15s %-15s %-15s %-15s\n", "BenchmarkObject_BuildComplex", complex_fson " ns/op", complex_fson_allocs " allocs/op", complex_fson_bytes " B/op", "1x (baseline)", "-"; \
		printf "%-30s %-15s %-15s %-15s %-15s %-15s\n", "BenchmarkJson_StdlibComplex", complex_std " ns/op", complex_std_allocs " allocs/op", complex_std_bytes " B/op", sprintf("%.2fx", complex_std/complex_fson), sprintf("%.2f%%", (complex_std-complex_fson)/complex_std*100); \
		printf "%-30s %-15s %-15s %-15s %-15s %-15s\n", "-----------------------------", "---------------", "---------------", "---------------", "---------------", "---------------"; \
	} \
	/BenchmarkObject_BuildLarge/ { large_fson = $$3; large_fson_allocs = $$5; large_fson_bytes = $$7; } \
	/BenchmarkJson_StdlibLarge/ { \
		large_std = $$3; large_std_allocs = $$5; large_std_bytes = $$7; \
		printf "%-30s %-15s %-15s %-15s %-15s %-15s\n", "BenchmarkObject_BuildLarge", large_fson " ns/op", large_fson_allocs " allocs/op", large_fson_bytes " B/op", "1x (baseline)", "-"; \
		printf "%-30s %-15s %-15s %-15s %-15s %-15s\n", "BenchmarkJson_StdlibLarge", large_std " ns/op", large_std_allocs " allocs/op", large_std_bytes " B/op", sprintf("%.2fx", large_std/large_fson), sprintf("%.2f%%", (large_std-large_fson)/large_std*100); \
		printf "%-30s %-15s %-15s %-15s %-15s %-15s\n", "-----------------------------", "---------------", "---------------", "---------------", "---------------", "---------------"; \
	} \
	END { \
		printf "\n%-30s %-15s %-15s %-15s\n", "Summary", "fson", "stdlib", "Improvement"; \
		printf "%-30s %-15s %-15s %-15s\n", "-----------------------------", "---------------", "---------------", "---------------"; \
		printf "%-30s %-15s %-15s %-15s\n", "Simple Case", simple_fson " ns/op", simple_std " ns/op", sprintf("%.2f%%", (simple_std-simple_fson)/simple_std*100); \
		printf "%-30s %-15s %-15s %-15s\n", "Complex Case", complex_fson " ns/op", complex_std " ns/op", sprintf("%.2f%%", (complex_std-complex_fson)/complex_std*100); \
		printf "%-30s %-15s %-15s %-15s\n", "Large Case", large_fson " ns/op", large_std " ns/op", sprintf("%.2f%%", (large_std-large_fson)/large_std*100); \
		printf "%-30s %-15s %-15s %-15s\n", "Average Improvement", "-", "-", sprintf("%.2f%%", ((simple_std-simple_fson)/simple_std + (complex_std-complex_fson)/complex_std + (large_std-large_fson)/large_std)*100/3); \
	}'
