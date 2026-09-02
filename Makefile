.PHONY: build test run templ templ-watch tailwind

build:
	@go tool templ generate
	@corepack pnpm run build
	@go build -o tmp/main .

test:
	@go test ./...

run: build
	@./tmp/main

templ:
	@go tool templ generate

templ-watch:
	@go tool templ generate --watch

tailwind:
	@corepack pnpm run css:dev
