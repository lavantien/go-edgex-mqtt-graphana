# syntax=docker/dockerfile:1

# build stage
FROM golang:buster AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstraper .

# deploy stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=build /app/admin/sensorClusterDeviceProfile.yaml ./
COPY --from=build /app/bootstraper ./bootstraper
COPY --from=build /app/web/ ./web/
COPY --from=build /app/common/ ./common/
EXPOSE 4321
ENTRYPOINT ["/bootstraper"]