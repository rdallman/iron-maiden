#! /bin/bash

bytes=(128 1024 32768 65480 262144)

for b in ${bytes[*]}
do
  ./iron-maiden "100000"    "1" "100" "1" "$b"
done
