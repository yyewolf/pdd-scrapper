FROM golang:latest as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
COPY --from=builder /app/main /app/main
WORKDIR /app
CMD ["/app/main"]