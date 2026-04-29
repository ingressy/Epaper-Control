FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Europe/Berlin

WORKDIR /root/
COPY --from=builder /app/main .

EXPOSE 70

CMD [ "./main" ]
