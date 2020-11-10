FROM golang:1.13

WORKDIR /app/lenslocked.com

COPY . .

RUN go get ./...

RUN go build

CMD ["./lenslocked.com"]
