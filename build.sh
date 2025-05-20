#!/bin/sh

mkdir -p build
env GOOS=linux GOARCH=amd64 go build -o build/notrock2tg_linux_amd64
env GOOS=windows GOARCH=amd64 go build -o build/notrock2tg_windows_amd64
env GOOS=darwin GOARCH=amd64 go build -o build/notrock2tg_darwin_amd64
env GOOS=darwin GOARCH=arm64 go build -o build/notrock2tg_darwin_arm64