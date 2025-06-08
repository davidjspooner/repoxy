#!/bin/bash

set -xeuo pipefail

export CMD=repoxy

TMP_ENV_FILE=$(mktemp)
ci-utility git suggest-build-env > "$TMP_ENV_FILE"
if [[ ! -s "$TMP_ENV_FILE" ]]; then
    echo "Failed to determine build environment"
    rm -f "$TMP_ENV_FILE"
    exit 1
fi

# check that the environment file is valid
if ! grep -qE '^BUILD_TIME=".+"$' "$TMP_ENV_FILE"; then
    echo "Invalid environment file format: $TMP_ENV_FILE"
    cat -vet "$TMP_ENV_FILE"
    #rm -f "$TMP_ENV_FILE"
    exit 1
fi

source "$TMP_ENV_FILE"
echo -e "Using build environment:\n$(cat "$TMP_ENV_FILE")\n"
rm -f "$TMP_ENV_FILE"

ci-utility matrix run \
    -d GOOS=linux \
    -d GOARCH=amd64,arm64 \
    -d CGO_ENABLED=0 \
    -- \
    go build -ldflags="-s -w \
        -X 'github.com/davidjspooner/go-text-cli/pkg/cmd.BUILD_VERSION=${BUILD_VERSION}' \
        -X 'github.com/davidjspooner/go-text-cli/pkg/cmd.BUILD_BY=${BUILD_BY}' \
        -X 'github.com/davidjspooner/go-text-cli/pkg/cmd.BUILD_TIME=${BUILD_TIME}' \
        -X 'github.com/davidjspooner/go-text-cli/pkg/cmd.BUILD_FROM=${BUILD_FROM}' " \
        -o 'dist/${CMD}-${GOOS}-${GOARCH}' './cmd/${CMD}'

# OS=$(uname | tr '[:upper:]' '[:lower:]')
# ARCH=$(uname -m)
# case "$ARCH" in
#     x86_64) ARCH="amd64" ;;
#     aarch64) ARCH="arm64" ;;
#     *) echo "Unsupported architecture: ${ARCH}"; exit 1 ;;
# esac
# 
# echo -e "\nCopying binary to /usr/local/bin/${CMD}\n"
# sudo rm -f /usr/local/bin/${CMD}
# sudo cp dist/${CMD}-"${OS}"-"${ARCH}" /usr/local/bin/${CMD}
# /usr/local/bin/${CMD} version