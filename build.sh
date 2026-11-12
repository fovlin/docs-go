#!/bin/bash
set -e

mkdir -p build
rm -rf build/docs-go_*

export CGO_ENABLED=0

export GOARCH=amd64
export GOOS=linux
go build -trimpath -o build/docs-go_linux_amd64/docs-go .

export GOOS=windows
go build -trimpath -o build/docs-go_windows_amd64/docs-go.exe .

export GOOS=darwin
go build -trimpath -o build/docs-go_darwin_amd64/docs-go .

export GOARCH=arm64

export GOOS=linux
go build -trimpath -o build/docs-go_linux_arm64/docs-go .

export GOOS=windows
go build -trimpath -o build/docs-go_windows_arm64/docs-go.exe .

export GOOS=darwin
go build -trimpath -o build/docs-go_darwin_arm64/docs-go .

for dir in build/docs-go_*; do
	cp -R theme "$dir/"
done

cd build

names=$(ls)

for name in ${names}; do
	tar -czf "${name}.tar.gz" "${name}"
done

for name in ${names}; do
	rm -rf "${name}"
done
