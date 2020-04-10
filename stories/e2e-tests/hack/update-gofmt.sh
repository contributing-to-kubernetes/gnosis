#!/usr/bin/env bash

find_files() {
  find . -not \( \
      \( \
        -wholename './output' \
        -o -wholename './.git' \
        -o -wholename './_output' \
        -o -wholename './_gopath' \
      \) -prune \
    \) -name '*.go'
}

find_files | xargs gofmt -s -w
