#!/bin/sh
# Generate test coverage statistics for Go packages.
#
# Works around the fact that `go test -coverprofile` currently does not work
# with multiple packages, see https://code.google.com/p/go/issues/detail?id=6909
#

workdir=.cover
profile="$workdir/cover.out"
mode=count


# check for failures in the test results file
# only look failures not equal 0, if <testsuite tests="14" failures="1" time="0.007" name="github.com/gaia-docker/...">

checkFailures() {
  grep -F "testsuite " $1
  if [ $? -eq 0 ]; then
    grep -F "failures=\"0\"" $1
    if [ $? -ne 0 ]; then
      testsuite_failures="$testsuite_failures""$1"";"
    fi
  fi
}

generate_cover_data() {
  rm -rf "$workdir"
  mkdir "$workdir"

  for pkg in "$@"; do
    f="$workdir/$(echo $pkg | tr / -).cover"
    tf="$workdir/$(echo $pkg | tr / -)_tests.xml"
    go test -v -covermode="$mode" -coverprofile="$f" "$pkg" | go-junit-report > "$tf"
    checkFailures "$tf"
  done

  echo "mode: $mode" >"$profile"
  grep -h -v "^mode:" "$workdir"/*.cover >>"$profile"

  if [ -z "$testsuite_failures" ]; then
    exit 0
  else
    echo "FAILED TESTSUITE(S) FOUND: $testsuite_failures"
    exit 13
  fi
}

generate_cover_data $(go list ./... | grep -v vendor)