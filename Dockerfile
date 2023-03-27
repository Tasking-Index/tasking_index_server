FROM golang:latest

WORKDIR /usr/src/app
COPY . .
RUN go mod tidy
WORKDIR /usr/src/app/api
EXPOSE 3000
CMD go run api.go
