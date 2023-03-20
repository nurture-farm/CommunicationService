#!/bin/bash

build_out=$(sh build.sh $1 | tee /dev/fd/2 | grep "Successfully built ")
if [[ -z "$build_out" ]]
then
  echo "Exiting since build failed"
  exit 1
fi

parsed_tag=""
if [[ $build_out =~ Successfully[[:space:]]built[[:space:]]([a-z0-9]+)$ ]]
then
  parsed_tag=${BASH_REMATCH[1]}
fi
echo "Parsed tag is $parsed_tag"