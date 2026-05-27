.PHONY: backend-tidy backend-fmt backend-test backend-run desktop-install desktop-dev desktop-build docker-build

backend-tidy:
	cd apps/backend && go mod tidy

backend-fmt:
	cd apps/backend && gofmt -w ./cmd ./internal

backend-test:
	cd apps/backend && go test ./...

backend-run:
	cd apps/backend && go run ./cmd/rayflowd

desktop-install:
	cd apps/desktop && npm install

desktop-dev:
	cd apps/desktop && npm run dev

desktop-build:
	cd apps/desktop && npm run build

docker-build:
	docker build -f apps/backend/Dockerfile -t rayflowd:dev .

build-sidecars:
	# Windows 64-bit Sidecar
	env GOOS=windows GOARCH=amd64 go build -o apps/desktop/src-tauri/bin/rayflowd-x86_64-pc-windows-msvc.exe apps/backend/cmd/rayflowd/main.go || go build -o apps/desktop/src-tauri/bin/rayflowd-x86_64-pc-windows-msvc.exe apps/backend/cmd/rayflowd/main.go
	# macOS Intel Sidecar
	env GOOS=darwin GOARCH=amd64 go build -o apps/desktop/src-tauri/bin/rayflowd-x86_64-apple-darwin apps/backend/cmd/rayflowd/main.go || true
	# macOS Apple Silicon Sidecar
	env GOOS=darwin GOARCH=arm64 go build -o apps/desktop/src-tauri/bin/rayflowd-aarch64-apple-darwin apps/backend/cmd/rayflowd/main.go || true
	# Linux 64-bit Sidecar
	env GOOS=linux GOARCH=amd64 go build -o apps/desktop/src-tauri/bin/rayflowd-x86_64-unknown-linux-gnu apps/backend/cmd/rayflowd/main.go || true


