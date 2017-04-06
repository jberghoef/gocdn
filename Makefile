default: requirements build

requirements:
	go get github.com/boltdb/bolt
	go get github.com/carlescere/scheduler
	go get gopkg.in/kyokomi/emoji.v1

build:
	go build *.go

reset:
	rm -rf cache cache.db
