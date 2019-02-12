#!/bin/bash
curl -d "stat=tuxbot - nightly - success&email=$STATHAT_EZKEY&value=1" http://api.stathat.com/ez
