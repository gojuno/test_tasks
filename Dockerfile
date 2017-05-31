FROM golang:alpine AS builder

ENV GOPATH /juno-test/
WORKDIR /juno-test/

COPY ./ /juno-test/
RUN apk --no-cache add git
RUN mkdir -p /juno-test/bin
RUN cd /juno-test/src/ && go get -v ./...
RUN cd /juno-test/cmd/ && go get -d ./...

# @todo #1:15m/QA uncomment this if you need high quality code or use ```docker-compose run linter```
# RUN go get -u github.com/alecthomas/gometalinter
# RUN /juno-test/bin/gometalinter --install
# RUN /juno-test/bin/gometalinter --config=/juno-test/.gometalinter.json --exclude=/go/src/ /juno-test/src/junoKvServer/... /juno-test/src/junoKvClient/...


RUN go build -o ./bin/kv_server ./cmd/kv_server.go
RUN go build -o ./bin/kv_client ./cmd/kv_client.go

RUN go test -c -o ./bin/kv_client_bench junoKvClient
RUN go test -c -o ./bin/kv_server_bench junoKvServer

# @todo #1:15m/QA uncomment this after resolve https://github.com/golang/go/issues/14481
# RUN apk --no-cache add gcc g++ musl musl-dev
# RUN go test -race -c -o ./bin/kv_server_bench_race juno_kv_client
# RUN go test -race -c -o ./bin/kv_server_bench_race juno_kv_server

FROM alpine:latest
COPY --from=builder /juno-test/bin/ /juno-test/bin/
EXPOSE 8379
ENTRYPOINT ["/bin/sh","-c"]


MAINTAINER Eugene Klimov <bloodjazman@gmail.com>
