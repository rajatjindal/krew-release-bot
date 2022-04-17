#! /bin/bash

VERSION=v0.0.43
docker run --rm -v `pwd`:/home/app rajatjindal/krew-release-bot:$VERSION krew-release-bot template 