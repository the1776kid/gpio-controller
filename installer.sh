#!/bin/bash
program_name="gpio-controller"
function build() {
    GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -v -ldflags="-s -w" -o ${program_name} .
}
function install_binary() {
    build
    install_binary
}
case $1 in
"build")
  build;;
"install")
install
  ;;
esac
build