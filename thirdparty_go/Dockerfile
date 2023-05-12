# syntax=docker/dockerfile:1

# build stage
FROM golang:buster AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o thirdparty .

# deploy stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=build /app/thirdparty ./thirdparty
EXPOSE 7070
ENTRYPOINT ["/thirdparty"]