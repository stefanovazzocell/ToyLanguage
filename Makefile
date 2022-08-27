.PHONY: run
run:
	go run ./cmd

.PHONY: build
build:
	go build -o tl ./cmd

.PHONY: bench
bench:
	# go test -bench . -run NoTest -cover ./cmd
	go test -bench . -run NoTest -cover ./src

.PHONY: test
test:
	# go test -race -cover ./cmd
	CGO_ENABLED=1 go test -race -cover ./src

.PHONY: testAnalysis
testAnalysis:
	# go test -race -cover -gcflags="-m" ./cmd
	go test -race -cover -gcflags="-m" ./src