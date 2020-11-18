FROM golang:1.13-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh && \
    apk add build-base

WORKDIR /app
