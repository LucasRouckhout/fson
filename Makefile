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
