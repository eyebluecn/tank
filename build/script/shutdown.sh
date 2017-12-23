#!/bin/bash


EXE_PATH=$GOPATH/bin/tank

EDASPID=`ps -ef | grep "$EXE_PATH"|grep -v grep |head -n 1 | awk '{print $2}'`
if [ -z $EDASPID ];
then
        echo "Cannot find $EXE_PATH."
else
        kill -9 $EDASPID
        echo $EXE_PATH
        echo 'Shutdown successfully.'
fi