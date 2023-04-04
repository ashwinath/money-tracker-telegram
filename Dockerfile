FROM golang:1.20-alpine as builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/app/money-tracker-telegram ./cmd/main.go

FROM alpine:3

ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO /zoneinfo.zip

WORKDIR /usr/src/app
COPY --from=builder /usr/local/app/money-tracker-telegram ./money-tracker-telegram

CMD ["./money-tracker-telegram"]
