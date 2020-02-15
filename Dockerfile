FROM golang:1.13.6-alpine3.11 as builder

ADD . /build

ARG LDFLAGS=""

ENV GOPROXY="https://goproxy.io"
ENV GO111MODULE="on"

WORKDIR /build
RUN go build -ldflags "${LDFLAGS}" -o aginx aginx.go


FROM nginx:1.17.7-alpine
MAINTAINER Haiker ni@renzhen.la

COPY --from=builder /build/aginx /usr/sbin/aginx

ENV AGINX_DEBUG="false"
ENV AGINX_LEVEL="info"
ENV AGINX_API=":8011"
ENV AGINX_SECURITY=""
ENV AGINX_CLUSTER=""
ENV AGINX_EXPOSE=""
ENV AGINX_WATCHER="false"

EXPOSE 8011

ENTRYPOINT ["/usr/sbin/aginx"]