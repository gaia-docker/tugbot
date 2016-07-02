#!/bin/sh
# Generate test coverage statistics for Go packages.
#
# Works around the fact that `go test -coverprofile` currently does not work
# with multiple packages, see https://code.google.com/p/go/issues/detail?id=6909
#
# Usage: script/coverage [--html]
#
#     --html  Additionally create HTML report and open it in browser
#

set -e

workdir=.cover
profile="$workdir/cover.out"
mode=count

generate_cover_data() {
  rm -rf "$workdir"
  mkdir "$workdir"

  for pkg in "$@"; do
    f="$workdir/$(echo $pkg | tr / -).cover"
    tf="$workdir/$(echo $pkg | tr / -)_tests.xml"
    go test -v -covermode="$mode" -coverprofile="$f" "$pkg" | go-junit-report > "$tf"
  done

  echo "mode: $mode" >"$profile"
  grep -h -v "^mode:" "$workdir"/*.cover >>"$profile"
}

show_cover_report() {
  go tool cover -${1}="$profile" -o "$workdir/coverage.html"
}

generate_cover_data $(go list ./... | grep -v vendor)

case "$1" in
  "")
    ;;
  --html)
    show_cover_report html ;;
  --help)
    echo >&2 "usage: $0 --coveralls|--html"; exit 0 ;;
  *)
    echo >&2 "error: invalid option: $1"; exit 1 ;;
esac
