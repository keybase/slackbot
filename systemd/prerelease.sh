#! /usr/bin/env bash

set -e -u -o pipefail

cd ~/client

git checkout -f master

git pull --ff-only

exec ./packaging/linux/docker_build.sh prerelease HEAD
