FROM golang:1.27 AS builder

WORKDIR /src

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/app .

ENTRYPOINT ["/app"]
