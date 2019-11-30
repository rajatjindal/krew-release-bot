#!/bin/bash

## update dependencies
dep ensure

## this cause openfaas deployment. probably due to a broken symlink
rm -rf vendor/sigs.k8s.io/krew/pkg/installation/testdata/