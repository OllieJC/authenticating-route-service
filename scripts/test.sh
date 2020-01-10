#!/bin/bash

HLP=$(pwd)
FILES=$(find -iname "*suite_test.go" | sed 's#\(.*\)/.*$#\1#')
for F in $FILES
do
  echo "Running go test in $F"
  cd $F
  go test
  if [ $? -ne 0 ]; then
    exit $?
  fi
  cd $HLP
done
