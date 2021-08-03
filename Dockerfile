FROM golang:1.16 as gopher

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./
COPY Makefile ./

RUN make compile

FROM alpine:latest as final

WORKDIR /app

COPY --from=gopher /app/bin/bootnode-registrar-linux-amd64 /app/bootnode-registrar

ENTRYPOINT [ "./bootnode-registrar" ]
EXPOSE 9898
