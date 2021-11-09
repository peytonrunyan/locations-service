# syntax=docker/dockerfile:1

FROM golang:1.16-alpine as builder
ENV GO111MODULE=on

WORKDIR /go/src/app

COPY . .
RUN go build ./cmd/geoservice

FROM alpine:3.11.3
COPY --from=builder /go/src/app/geoservice .
COPY .envDocker .

RUN mv .envDocker .env

WORKDIR /data
COPY ./data/us_states_500k.geojson .
COPY ./data/Merged_Counties_Subcounties_Communities.geojson .

WORKDIR /

EXPOSE 8083

CMD [ "./geoservice" ]