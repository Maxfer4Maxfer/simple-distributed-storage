FROM golang:alpine as build

RUN apk add ca-certificates 

WORKDIR /opt

COPY . . 

RUN go build -o bin/api-server cmd/api-server/main.go

#######################################

FROM alpine:latest

WORKDIR /opt

COPY --from=build /opt/bin/api-server .

ENTRYPOINT [ "/opt/api-server"]
