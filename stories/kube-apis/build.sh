#!/bin/bash

# Source directory, will be mounted to /src, defaults to ./tools.
SOURCE_DIR="${SOURCE_DIR:-$(pwd -P)/tools}"


docker run --rm \
  `# run as the user/group running this script` \
  --user "$(id -u):$(id -g)" \
  `# mount golang cache - required for 'go build'ing` \
  -v "${CACHE_VOLUME}:/go" -e XDG_CACHE_HOME=/go/cache \
  `# mount source directory` \
  -v "${SOURCE_DIR}:/src" \
  `# set environment variables to allow for cross-compilation (if desired)` \
  -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 \
  `# set working directory to the source directory` \
  -w "/src" \
  golang:1.14 "$@"
