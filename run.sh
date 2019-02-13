#!/bin/sh
rm -f pewnit
go build -o pewnit main.go 
./pewnit "$@"