FROM golang:1.18 AS go

ADD . /app
WORKDIR /app
RUN go build server.go

FROM ubuntu:22.04

RUN apt-get update && apt install -y bubblewrap
RUN apt-get install -y tint
COPY  --from=go /app/server /usr/bin/remote-login

CMD ["/usr/bin/remote-login", "--public", "bwrap", "--ro-bind", "/", "/", "--proc", "/proc", "--dev", "/dev", "--unshare-all", "--uid", "0", "/usr/games/tint"]
