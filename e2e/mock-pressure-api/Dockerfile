FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY e2e/mock-pressure-api/main.go ./main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/mock-pressure-api ./main.go

FROM alpine:3.20
COPY --from=builder /bin/mock-pressure-api /usr/local/bin/mock-pressure-api
ENTRYPOINT ["/usr/local/bin/mock-pressure-api"]
