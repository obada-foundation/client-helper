
FROM golang:1.21 as builder

WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download -x

COPY . .

RUN CGO_ENABLED=0 go build -a -ldflags '-s -X "google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn"'  -tags=badgerdb -o /client-helper main.go

FROM alpine:3.19

RUN apk add ca-certificates --no-cache && adduser -D obada

USER obada

WORKDIR /home/obada

COPY --from=builder --chown=obada /client-helper /home/obada

RUN mkdir /home/obada/keys

CMD ["/home/obada/client-helper", "server"]
