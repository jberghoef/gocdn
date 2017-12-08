all: reset install test build

build:
	go build *.go

test:
	go test -v

reset:
	rm -rf cache cache.db

install:
	go get -d -v
	go install -v

docker_build:
	docker build -t gocdn .

docker_run:
	docker run -it --rm --name gocdn gocdn
