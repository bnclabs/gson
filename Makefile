build:
	go build

test:
	go test -race -timeout 4000s -test.run=. -test.bench=xxx -test.benchmem=true
	go test -timeout 4000s -test.run=xxx -test.bench=. -test.benchmem=true

coverage:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm -rf coverage.out
