.PHONY: clean
clean:
	@rm -rf dist
	@go mod tidy

download:
	@go mod download

dist:
	@mkdir -p dist/bin

dist/bin/goscript: dist
	@go build -o dist/bin/goscript cmd/goscript/*.go

install-goscript:
	@go install tools/cmd/goscript
