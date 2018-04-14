FROM golang:alpine

WORKDIR /go/src/pulseengine
COPY . .

RUN apk add --no-cache git mercurial \
    && go get -d -v ./... \
    && apk del git mercurial

RUN go install -v ./...

CMD ["server"]