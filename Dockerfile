FROM golang:1.10.3

ENV PROTOCOL=https
ENV ORIGIN=www.example.com

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
RUN go build .

CMD ["app"]
