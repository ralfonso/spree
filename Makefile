spree_dir := $(abspath $(shell git rev-parse --show-toplevel))
build_image = "golang:1.7.1-alpine"

all: spreed spree-client
.PHONY: all

spreed:
	docker run --rm -v $(spree_dir):/go/src/github.com/ralfonso/spree \
		-w /go/src/github.com/ralfonso/spree \
		-it $(build_image) \
		go build -o spreed github.com/ralfonso/spree/cmd/spreed
.PHONY: spreed

spree-client:
	docker run --rm -v $(spree_dir):/go/src/github.com/ralfonso/spree \
		-w /go/src/github.com/ralfonso/spree \
		-it $(build_image) \
		go build -o spree-client github.com/ralfonso/spree/cmd/spree-client
.PHONY: spree-client

image-spreed:
	docker build -f Dockerfile.spreed -t ralfonso/spreed .

push-spreed:
	docker push ralfonso/spreed
