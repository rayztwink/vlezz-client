APP := rayflowd

.PHONY: tidy fmt test run build clean

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

run:
	go run ./cmd/rayflowd

build:
	go build -o bin/$(APP) ./cmd/rayflowd

clean:
	rm -rf bin data

