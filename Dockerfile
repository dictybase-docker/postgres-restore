FROM golang:alpine
MAINTAINER Siddhartha Basu <siddhartha-basu@northwestern.edu>
COPY . /usr/src/app
RUN cd /usr/src/app \
    && go-wrapper download \
    && go build \
    && cp app ${GOPATH}bin/


