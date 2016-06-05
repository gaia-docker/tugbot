FROM alpine:3.3

COPY .dist/tugbot /usr/bin/tugbot

LABEL gaiadocker.tugbot=true

ENTRYPOINT ["/usr/bin/tugbot"]