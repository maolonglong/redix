.PHONY: default build build_image test clean

default: build

build:
	@./scripts/build.sh

build_image:
	@./scripts/build_image.sh

test:
	go test -v -count=1 -race ./redix/...

clean:
	rm -rf ./bin/
	rm -rf ./data/
	rm -rf ./dist/
