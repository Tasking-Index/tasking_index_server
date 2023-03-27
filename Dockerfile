FROM golang:latest

WORKDIR /usr/src/app
COPY . .
RUN go mod download
WORKDIR /usr/src/app/api
EXPOSE 3000
CMD go run api.go
