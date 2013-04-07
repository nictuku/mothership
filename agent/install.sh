#!/bin/bash

set -eu

mkdir -p $HOME/myservers
cd $HOME/myservers
curl -sLO http://geekroot.com/static/agent
chmod 700 agent
cat <<EOF> run.sh
nohup ./agent 2> /dev/null &
echo $! > pid
EOF
chmod 700 run.sh
./run.sh