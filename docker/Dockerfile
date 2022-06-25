FROM alpine:edge AS builder

RUN apk update && apk upgrade && apk add go

COPY ./src /src

WORKDIR /src

RUN CGO_ENABLED=0 go build -o /client-helper main.go

FROM alpine:3.14

RUN apk add ca-certificates --no-cache && adduser -D obada

USER obada

WORKDIR /home/obada

COPY --from=builder --chown=obada /client-helper /home/obada

RUN mkdir /home/obada/keys

CMD ["/home/obada/client-helper", "server"]
