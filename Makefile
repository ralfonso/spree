spree_dir := $(abspath $(shell git rev-parse --show-toplevel))
build_image = "golang:1.8-alpine"

all: spreed spreectl
.PHONY: all

spreed: bindata
	docker run --rm -v $(spree_dir):/go/src/github.com/ralfonso/spree \
		-w /go/src/github.com/ralfonso/spree \
		-it $(build_image) \
		go build -o spreed github.com/ralfonso/spree/cmd/spreed
.PHONY: spreed

spreectl:
	docker run --rm -v $(spree_dir):/go/src/github.com/ralfonso/spree \
		-w /go/src/github.com/ralfonso/spree \
		-it $(build_image) \
		go build -o spreectl github.com/ralfonso/spree/cmd/spreectl
.PHONY: spreectl

spreectl-native:
	go build -o spreectl github.com/ralfonso/spree/cmd/spreectl
.PHONY: spreectl-native

image:
	docker build -f Dockerfile.spreed -t ralfonso/spreed .

push:
	docker push ralfonso/spreed

bindata:
	go-bindata -o cmd/spreed/bindata.go static/...
