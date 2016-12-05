FROM golang:1.7.1-alpine

# install required libs
RUN apk --no-cache add git bash curl

 # install glide package manager
RUN curl -Ls https://github.com/Masterminds/glide/releases/download/v0.12.1/glide-v0.12.1-linux-amd64.tar.gz | tar xz -C /tmp \
&& mv /tmp/linux-amd64/glide /usr/bin/

# gox - Go cross compile tool
RUN go get -v github.com/mitchellh/gox

# cover - Go code coverage tool
RUN go get -v golang.org/x/tools/cmd/cover

# go-junit-report - convert Go test into junit.xml format
RUN go get -v github.com/jstemmer/go-junit-report

CMD ["script/go_build.sh"]
