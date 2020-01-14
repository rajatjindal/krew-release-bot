#!/bin/bash

version=$1
if [ "$version" == "" ]; then
	echo "version not provided"
	exit 1
fi

if [ "$PROJECT_ID" == "" ]; then 
	echo "cloud run project id not provided"
	exit 1
fi

## push for github actions
docker build . -t rajatjindal/krew-release-bot:$version -f action.Dockerfile
docker push rajatjindal/krew-release-bot:$version

## push for cloud run
docker build . -t gcr.io/$PROJECT_ID/krew-release-bot:$version -f webhook.Dockerfile
docker push gcr.io/$PROJECT_ID/krew-release-bot:$version
