FROM golang:1.13.6-alpine3.11 as builder
ADD . /build
WORKDIR /build
RUN go build -ldflags "$BUILD_DEBUG $BUILD_PARAM" -o aginx aginx.go


FROM nginx:1.8.1-alpine
WORKDIR /apps
COPY --from=builder /build/aginx /usr/sbin/aginx
VOLUME /etc/nginx

ENV AGINX_API=""
ENV AGINX_SECURITY=""
ENV AGINX_CLUSTER=""

EXPOSE 80 8011

ENTRYPOINT ["/usr/sbin/aginx"]