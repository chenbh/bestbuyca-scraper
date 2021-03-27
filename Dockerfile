FROM golang:alpine as builder

WORKDIR /src
COPY go.* /src/
RUN go mod download

COPY . /src
RUN go build -o /scraper .

FROM alpine:3
COPY --from=builder /scraper /scraper

ENTRYPOINT /scraper
