all: reset install test build

build:
	go build -o gocdn *.go

test:
	go test -v

reset:
	rm -rf cache cache.db

install:
	dep ensure

docker_build:
	docker build -t gocdn .

docker_run:
	docker run -it --rm --name gocdn gocdn
