FROM golang:1.18-alpine as builder
WORKDIR /app
RUN apk update && apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    && update-ca-certificates

COPY . .
RUN go mod download
RUN go mod verify
WORKDIR /app/cmd/server
RUN GOOS=linux GOARCH=amd64 go build -mod vendor -a -installsuffix cgo -o z4

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app/cmd/server/z4 /usr/local/bin/z4
ENTRYPOINT ["z4"]
EXPOSE 6355 6356 3306