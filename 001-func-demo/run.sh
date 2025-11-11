#!/usr/bin/env bash

source $(git rev-parse --show-toplevel)/lib.sh

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $SCRIPT_DIR

# Cleanup old demo
rm -rf square

bx func create -l go square
bx tree -C square
bx bat square/handle.go

cp handle.go square/handle.go

bx bat square/handle.go

pushd square >/dev/null
  bx func build --registry dprotaso
  x func run &
  read
  bx func invoke --data '{"value":"9"}'
  bx func deploy
  bx func invoke --insecure --data '{"value":"9"}'
popd >/dev/null

kill %1

bx func create -h | bat
kubectl delete ksvc --all > /dev/null
