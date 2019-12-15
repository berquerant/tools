.PHONY: clean
clean: rm-generated
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

go-stringer:
	@go get golang.org/x/tools/cmd/stringer

.PHONY: generate
generate:
	@go generate ./pkg/...

.PHONY: rm-generated
rm-generated:
	@find . -type f -name "*generated\.*\.go" | xargs rm
