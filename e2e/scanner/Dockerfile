FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/port-scan ./cmd/port-scan

FROM alpine:3.20
COPY --from=builder /bin/port-scan /usr/local/bin/port-scan
ENTRYPOINT ["/usr/local/bin/port-scan"]
