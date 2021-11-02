#!/bin/bash

export TAG=`git describe --tags`
echo TAG is $TAG


export GOOS=windows 
export GOARCH=amd64 
echo Building $GOOS $GOARCH
go build -o trh-$GOOS-$GOARCH-$TAG.exe .
zip trh-$GOOS-$GOARCH-$TAG.zip trh-$GOOS-$GOARCH-$TAG.exe LICENSE.md


export GOOS=linux 
export GOARCH=amd64 
echo Building $GOOS $GOARCH
go build -o  trh-$GOOS-$GOARCH-$TAG .
zip trh-$GOOS-$GOARCH-$TAG.zip trh-$GOOS-$GOARCH-$TAG LICENSE.md


export GOOS=darwin 
export GOARCH=arm64 
echo Building $GOOS $GOARCH
go build -o trh-$GOOS-$GOARCH-$TAG .
zip trh-$GOOS-$GOARCH-$TAG.zip trh-$GOOS-$GOARCH-$TAG LICENSE.md


export GOOS=darwin 
export GOARCH=amd64 
echo Building $GOOS $GOARCH
go build -o trh-$GOOS-$GOARCH-$TAG .
zip trh-$GOOS-$GOARCH-$TAG.zip trh-$GOOS-$GOARCH-$TAG LICENSE.md


mkdir -p ../../../ejfhp/website/static/binaries/trh/$TAG
mv trh-*.zip ../../../ejfhp/website/static/binaries/trh/$TAG
rm trh-*
