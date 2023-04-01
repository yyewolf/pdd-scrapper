FROM golang:alpine3.17 as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o scrapper .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/scrapper /app/scrapper
RUN chmod +x scrapper && mkdir /app/out
CMD [ "/app/scrapper" ]