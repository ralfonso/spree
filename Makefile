spree_dir := $(abspath $(shell git rev-parse --show-toplevel))
build_image = "golang:1.7.1-alpine"

all: spreed spreectl
.PHONY: all

spreed: server-bindata
	docker run --rm -v $(spree_dir):/go/src/github.com/ralfonso/spree \
		-w /go/src/github.com/ralfonso/spree \
		-it $(build_image) \
		go build -o spreed github.com/ralfonso/spree/cmd/spreed
.PHONY: spreed

spreectl: client-bindata
	docker run --rm -v $(spree_dir):/go/src/github.com/ralfonso/spree \
		-w /go/src/github.com/ralfonso/spree \
		-it $(build_image) \
		go build -o spreectl github.com/ralfonso/spree/cmd/spreectl
.PHONY: spreectl

spreectl-native: client-bindata
	go build -o spreectl github.com/ralfonso/spree/cmd/spreectl
.PHONY: spreectl-native

image:
	docker build -f Dockerfile.spreed -t ralfonso/spreed .

push:
	docker push ralfonso/spreed

server-bindata:
	go-bindata -o cmd/spreed/bindata.go private/server/... private/shared/...

client-bindata:
	go-bindata -o cmd/spreectl/bindata.go private/client/... private/shared/...
