#!/bin/bash

version=$1
if [ "$version" == "" ]; then
	echo "version not provided"
	exit 1
fi

## push for github actions
docker build . -t rajatjindal/krew-release-bot:$version -f Dockerfile
docker push rajatjindal/krew-release-bot:$version
