FROM golang:1.27-trixie AS build

WORKDIR /src

ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -trimpath -o=bootstrap -v 'cmd/main.go'

FROM gcr.io/distroless/base-debian13:latest

COPY --from=build /src/bootstrap /

CMD ["/bootstrap"]