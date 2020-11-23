FROM golang:1.15.5-alpine3.12 as builder

ENV GO111MODULE="on"
ARG LDFLAGS=""

ADD . /aginx
WORKDIR /aginx

RUN apk add --no-cache make build-base
RUN go build -tags bindata -ldflags "${LDFLAGS}" -o aginx aginx.go

FROM axizdkr/tengine:2.3.2
#FROM nginx:1.19.4-alpine

MAINTAINER Haiker ni@renzhen.la

COPY --from=builder /aginx/aginx /bin/aginx
ADD core/config/aginx.conf.example /etc/aginx/aginx.conf.example

RUN touch /etc/aginx/aginx.conf && \
    rm -f /etc/nginx/modules && \
    rm -rf /var/log/nginx/error.log /var/log/nginx/access.log && \
    ln -s /dev/stderr /var/log/nginx/error.log && \
    ln -s /dev/stdout /var/log/nginx/access.log

ENV AGINX_CONF=/etc/aginx/aginx.conf
ENV AGINX_PLUGINS=/usr/lib/aginx/modules
ENV AGINX_BACKUP_DIR=/var/lib/aginx/backups

VOLUME /etc/nginx
VOLUME /etc/aginx

EXPOSE 8011 80 443
ENTRYPOINT ["/bin/aginx"]
