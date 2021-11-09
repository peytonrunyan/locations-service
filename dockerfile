# syntax=docker/dockerfile:1

FROM golang:1.16-alpine
ENV GO111MODULE=on

WORKDIR /go/src/app

COPY . .
RUN mv .envDocker .env
RUN go build ./cmd/geoservice

EXPOSE 8083

CMD [ "./geoservice" ]