.PHONY: build test run templ templ-watch tailwind

build:
	@go tool templ generate
	@tailwindcss -i ./views/css/styles.css -o ./public/styles.css --minify
	@go build -o ./bin/main .

test:
	@go test ./...

air:
	@air --build.cmd "go build -o ./bin/main ." --build.entrypoint "./bin/main"

templ:
	@go tool templ generate --watch

tailwind:
	@tailwindcss -i ./views/css/styles.css -o ./public/styles.css --watch
