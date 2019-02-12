#!/bin/bash
curl -X POST -d "stat=tuxbot - nightly - success&ezkey=$STATHAT_EZKEY&count=1" https://api.stathat.com/ez
