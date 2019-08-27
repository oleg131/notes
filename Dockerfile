FROM golang:alpine
RUN mkdir /app
ADD main.go index.html /app/
WORKDIR /app
RUN apk add --no-cache git \
    && go get github.com/lib/pq \
    && apk del git
RUN go build -o main .
CMD ["/app/main"]
