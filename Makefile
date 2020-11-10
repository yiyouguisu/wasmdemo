.DEFAULT_GOAL := build

.PHONY: build

build:
	docker run -it -w /tmp/proxy-wasm-go -v $(shell pwd):/tmp/proxy-wasm-go tinygo/tinygo-dev:latest \
		tinygo build -o /tmp/proxy-wasm-go/${name}/main.go.wasm -scheduler=none -target=wasi \
		-wasm-abi=generic /tmp/proxy-wasm-go/${name}/main.go