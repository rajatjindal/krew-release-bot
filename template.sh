#! /bin/bash

VERSION=v0.0.39
docker run --rm -v `pwd`:/home/app rajatjindal/krew-release-bot:$VERSION krew-release-bot template 