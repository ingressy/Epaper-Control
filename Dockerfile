FROM golang:1.27 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .
RUN swag init -g main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o token ./rest/token

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Europe/Berlin

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/token .

EXPOSE 80

CMD [ "./main" ]
