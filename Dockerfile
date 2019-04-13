FROM golang:alpine

MAINTAINER "Adigun Hammed, Olalekan <adigunhammed.lekan@gmail.com>"
WORKDIR /go/src

COPY . /go/src

ENV GO111MODULE=on

RUN apk update && \
    apk upgrade && \
    apk add git

RUN go get .
RUN cd /go/src && go build -o main

EXPOSE 9008
ENTRYPOINT ["./main"]