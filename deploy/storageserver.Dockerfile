FROM golang:alpine as build

RUN apk add ca-certificates 

WORKDIR /opt

COPY . . 

RUN go build -o bin/storage-server cmd/storage-server/main.go

#######################################

FROM alpine:latest

WORKDIR /opt

COPY --from=build /opt/bin/storage-server .
RUN mkdir /opt/data

ENTRYPOINT [ "/opt/storage-server"]
