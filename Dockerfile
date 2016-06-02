FROM alpine:3.3

COPY .dist/mr-burns /usr/bin/tugbot

ENTRYPOINT ["/usr/bin/tugbot"]