build:
	go build -tags n1ql -o gson-n1ql
	go build -tags n1ql -o gson

test:
	go build -o gson
	go build -o gsonn1ql
	go test -race -test.run=. -test.bench=. -test.benchmem=true
	sort.sh
	transforms.sh
	rm -rf gson gsonn1ql
