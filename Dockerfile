FROM alpine:3.3

COPY .dist/tugbot /usr/bin/tugbot

ENTRYPOINT ["/usr/bin/tugbot"]