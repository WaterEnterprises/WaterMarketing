.PHONY: all build build-all run serve run-direct clean
.PHONY: web web-install web-dev web-build web-preview

BINARY=crm

all: build web-build

build:
	go build -o $(BINARY).exe ./cmd

run: build
	.\$(BINARY).exe $(ARGS)

serve:
	.\$(BINARY).exe serve

run-direct:
	go run ./cmd $(ARGS)

web web-install:
	cd web && bun install

web-dev:
	cd web && bun run dev

web-build:
	cd web && bun run build

web-preview:
	cd web && bun run preview

clean:
	rm -f $(BINARY).exe
	rm -rf web/build web/.svelte-kit
	go clean
