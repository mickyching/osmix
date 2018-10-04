#!/bin/bash

unset GO15VENDOREXPERIMENT
godep restore
rm -rf Godeps
export GO15VENDOREXPERIMENT=1
godep save
git add -A .
git commit -am "Godep workspace -> vendor/"

