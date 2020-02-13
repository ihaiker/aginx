FROM golang:1.13.6-alpine3.11 as builder

ADD . /build

ARG LDFLAGS=""

ENV GOPROXY="https://goproxy.io"
ENV GO111MODULE="on"

WORKDIR /build
RUN go build -ldflags "$LDFLAGS" -o aginx aginx.go


FROM nginx:1.8.1-alpine
MAINTAINER Haiker ni@renzhen.la

COPY --from=builder /build/aginx /usr/sbin/aginx
VOLUME /etc/nginx

ADD html /usr/share/nginx/

ENV AGINX_API=":8011"
ENV AGINX_SECURITY=""
ENV AGINX_CLUSTER=""
ENV AGINX_EXPOSE=""

EXPOSE 8011

ENTRYPOINT ["/usr/sbin/aginx"]