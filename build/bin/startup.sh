#!/bin/bash

# executable path
DIR="$( cd "$( dirname "$0"  )" && pwd  )"
EXE_PATH=$GOPATH/bin/tank

if [ -f "$EXE_PATH" ]; then
 nohup $EXE_PATH >/dev/null 2>&1 &
else
 echo 'Cannot find $EXE_PATH.'
 exit 1
fi
