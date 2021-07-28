#!/bin/bash

export TAG=`git describe --tags`
echo TAG is $TAG


export GOOS=windows 
export GOARCH=amd64 
echo Building $GOOS $GOARCH
go build -o trh-$GOOS-$GOARCH-$TAG.exe .


export GOOS=linux 
export GOARCH=amd64 
echo Building $GOOS $GOARCH
go build -o  trh-$GOOS-$GOARCH-$TAG .


export GOOS=darwin 
export GOARCH=arm64 
echo Building $GOOS $GOARCH
go build -o trh-$GOOS-$GOARCH-$TAG .


export GOOS=darwin 
export GOARCH=amd64 
echo Building $GOOS $GOARCH
go build -o trh-$GOOS-$GOARCH-$TAG .


mkdir  mkdir /home/diego/Code/ejfhp/ejfhp/website/static/binaries/trh/$TAG
mv trh-* /home/diego/Code/ejfhp/ejfhp/website/static/binaries/trh/$TAG
