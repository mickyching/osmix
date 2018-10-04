#!/bin/bash

echo go get github.com/rogpeppe/godef
go get github.com/rogpeppe/godef
echo go get github.com/nsf/gocode
go get github.com/nsf/gocode
echo go get github.com/tools/godep
go get github.com/tools/godep

mkdir -p $GOPATH/src/golang.org/x
cd $GOPATH/src/golang.org/x

echo git clone https://github.com/golang/net
git clone https://github.com/golang/net
echo go install golang.org/x/net
go install golang.org/x/net

echo git clone https://github.com/golang/tools
git clone https://github.com/golang/tools
echo go install golang.org/x/tools/cmd/...
go install golang.org/x/tools/cmd/...
