.PHONY: default build

default: build

build:
	@./scripts/build.sh

build_image:
	@./scripts/build_image.sh

test:
	go test -v -count=1 -race ./...

clean:
	rm -rf ./bin/
	rm -rf ./data/
