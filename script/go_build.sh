#!/bin/bash
distdir=.dist

go_build() {
  rm -rf "${distdir}"
  mkdir "${distdir}"
  go build
  CGO_ENABLED=0 go build -v -o ${distdir}/tugbot
}

go_build
