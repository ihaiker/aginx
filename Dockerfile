FROM golang:1.13.6-alpine3.11 as builder

ADD . /build
ENV GOPROXY="https://goproxy.io"
ENV GO111MODULE="on"
ARG LDFLAGS=""

RUN apk add --no-cache make build-base
RUN go build -ldflags "${LDFLAGS}" -o aginx aginx.go


FROM nginx:1.17.7-alpine
MAINTAINER Haiker ni@renzhen.la

COPY --from=builder /build/aginx /usr/sbin/aginx
COPY --from=builder /build/conf/aginx.conf /etc/nginx/aginx.conf

ENV AGINX_EMAIL="aginx@renzhen.la"
ENV AGINX_DEBUG="false" AGINX_LEVEL="info"

ENV AGINX_CONF="" AGINX_API="127.0.0.1:8011" AGINX_SECURITY=""
ENV AGINX_STORAGE="" AGINX_DISABLE_WATCHER=""
ENV AGINX_EXPOSE=""

ENV AGINX_DOCKER="" AGINX_DOCKER_API_VERSION="" AGINX_DOCKER_HOST="" AGINX_DOCKER_TLS_VERIFY="" AGINX_DOCKER_CERT_PATH=""
ENV AGINX_CONSUL="" AGINX_CONSUL_HTTP_ADDR="" AGINX_CONSUL_HTTO_TOKEN=""

EXPOSE 8011

CMD ["server"]
ENTRYPOINT ["/usr/sbin/aginx"]