#!/bin/sh

homePath=$GOPATH/src/tank

oldPath=$(pwd)

echo "cd homePath"
cd $homePath

echo "shutdown tank"
source $homePath/doc/script/shutdown.sh

echo "git reset"
git reset --hard HEAD

echo "git pull"
git pull

echo "go install tank"
go install tank

cd $oldPath

echo "startup tank"
source $homePath/doc/script/startup.sh


