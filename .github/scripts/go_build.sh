#!/bin/bash

go_build() {
    # Initialize variables
    path=""
    zip_name=""

    # Parse command-line arguments
    while [[ "$#" -gt 0 ]]; do
        case $1 in
            --path) path="$2"; shift ;;
            --zip) zip_name="$2"; shift ;;
            *) echo "Unknown parameter passed: $1"; exit 1 ;;
        esac
        shift
    done

    # Check values
    if [ -z "$path" ]; then
        echo "Path is required"
        exit 1
    fi

    if [ -z "$zip_name" ]; then
        echo "Zip name is required"
        exit 1
    fi

    # export variables
    boostrap_path=bin/$path/bootstrap
    export GO111MODULE=on
    export CGO_ENABLED=1
    export GOARCH=arm64
    # export GOOS=linux

    printf "Building $path\n"
    printf "Bootstrap path: $boostrap_path\n"
    printf "Zip name: $zip_name\n"
    # exec build and zip
    go build -ldflags="-s -w" -o $boostrap_path $path/main.go && \
    chmod +x $boostrap_path && \
    zip -j bin/$zip_name.zip $boostrap_path
}