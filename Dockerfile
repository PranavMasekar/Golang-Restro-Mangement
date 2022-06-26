FROM golang:1.16-alpine

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go mod download

RUN go build -o go-restaurant .

EXPOSE 8000

CMD ["/app/go-restaurant"]