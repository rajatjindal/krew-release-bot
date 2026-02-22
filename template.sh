#! /bin/bash

VERSION=v0.0.48
docker run --rm -v `pwd`:/home/app ghcr.io/rajatjindal/krew-release-bot:$VERSION krew-release-bot template 